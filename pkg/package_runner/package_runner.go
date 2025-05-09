package packagerunner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	util "github.com/JackBekket/reflexia/internal"
	store "github.com/JackBekket/reflexia/pkg"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/schema"
)

type PackageRunnerService struct {
	PkgFiles          map[string][]string
	ProjectConfig     *project.ProjectConfig
	SummarizeService  *summarize.SummarizeService
	EmbeddingsService *store.EmbeddingsService
	ExactPackages     *string
	OverwriteReadme   bool
	WithFileSummary   bool
}

func (s *PackageRunnerService) RunPackages() ([]string, []string, []string, []string, error) {
	fallbackFileResponses := []string{}
	emptyFileResponses := []string{}
	fallbackPackageResponses := []string{}
	emptyPackageResponses := []string{}

	for pkg, files := range s.PkgFiles {
		if *s.ExactPackages != "" &&
			!slices.Contains(strings.Split(*s.ExactPackages, ","), pkg) {
			continue
		}
		fmt.Printf("\nPackage %s\n", pkg)

		pkgFileMap := map[string]string{}
		pkgDir := ""
		for _, relPath := range files {
			fmt.Println(relPath)
			content, err := os.ReadFile(filepath.Join(s.ProjectConfig.RootPath, relPath))
			if err != nil {
				return nil, nil, nil, nil, err
			}

			contentStr := string(content)
			codeSummaryContent := "Empty file"

			if strings.TrimSpace(contentStr) != "" {
				codeSummaryContent, err = s.SummarizeService.LLMRequest(
					"%s```%s```",
					s.ProjectConfig.CodePrompt, contentStr,
				)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				if strings.TrimSpace(codeSummaryContent) == "" &&
					s.ProjectConfig.CodePromptFallback != nil {
					fallbackFileResponses = append(
						fallbackFileResponses,
						filepath.Join(s.ProjectConfig.RootPath, relPath),
					)
					codeSummaryContent, err = s.SummarizeService.LLMRequest(
						"%s```%s```",
						*s.ProjectConfig.CodePromptFallback, contentStr,
					)
					if err != nil {
						return nil, nil, nil, nil, err
					}
				}
			} else {
				fmt.Println("Empty file")
			}

			fmt.Printf("\n")
			pkgFileMap[relPath] = codeSummaryContent
			if s.EmbeddingsService != nil {
				ids, err := s.EmbeddingsService.Store.AddDocuments(
					context.Background(),
					[]schema.Document{
						{
							PageContent: string(content),
							Metadata: map[string]interface{}{
								"package":  pkg,
								"filename": relPath,
								"type":     "code",
							},
						}, {
							PageContent: codeSummaryContent,
							Metadata: map[string]interface{}{
								"package":  pkg,
								"filename": relPath,
								"type":     "doc",
							},
						},
					},
				)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				fmt.Printf("Succesfully pushed docs %s into embeddings vector store\n", ids)
			}

			if strings.TrimSpace(codeSummaryContent) == "" {
				emptyFileResponses = append(
					emptyFileResponses,
					filepath.Join(s.ProjectConfig.RootPath, relPath),
				)
			}
			if pkgDir == "" {
				pkgDir = filepath.Join(s.ProjectConfig.RootPath, filepath.Dir(relPath))
			}
		}
		if pkgDir == "" {
			log.Info().Msgf("There is no mathcing files for pkg: %s", pkg)
			continue
		}

		fileStructure, err := getDirFileStructure(pkgDir)
		if err != nil {
			log.Warn().Err(err).Msg("pkg/package_runner/package_runner.go:125: getDirFileStructure error")
		}

		fmt.Printf("\nSummary for a package %s: \n", pkg)
		// Generate Summary for a package (summarizing and group file summarization by package name)
		// Get whole map of code summaries as a string and toss it to summarize .MD for a package
		pkgSummaryContent, err := s.SummarizeService.LLMRequest(
			"%s\n\n%s\n%s",
			s.ProjectConfig.PackagePrompt,
			fileStructure,
			fileMapToString(pkgFileMap),
		)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if strings.TrimSpace(pkgSummaryContent) == "" && s.ProjectConfig.PackagePromptFallback != nil {
			fallbackPackageResponses = append(fallbackPackageResponses, pkg)
			pkgSummaryContent, err = s.SummarizeService.LLMRequest(
				"%s\n\n%s\n%s",
				*s.ProjectConfig.PackagePromptFallback,
				fileStructure,
				fileMapToString(pkgFileMap),
			)
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}

		if strings.TrimSpace(pkgSummaryContent) == "" {
			emptyPackageResponses = append(emptyPackageResponses, pkg)
		}

		if s.EmbeddingsService != nil {
			ids, err := s.EmbeddingsService.Store.AddDocuments(
				context.Background(),
				[]schema.Document{
					{
						PageContent: pkgSummaryContent,
						Metadata: map[string]interface{}{
							"package": pkg,
							"type":    "doc",
						},
					},
				},
			)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			fmt.Printf("Succesfully pushed docs %s into embeddings vector store\n", ids)
		}

		readmeFilename := "README.md"
		if !s.OverwriteReadme {
			readmeFilename, err = getReadmePath(pkgDir)
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}
		if err := writeFile(
			filepath.Join(pkgDir, readmeFilename),
			pkgSummaryContent,
		); err != nil {
			return nil, nil, nil, nil, err
		}

		if s.WithFileSummary {
			if err := writeFile(
				filepath.Join(pkgDir, "FILES.md"),
				fileMapToMd(pkgFileMap),
			); err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}
	return fallbackFileResponses, emptyFileResponses, fallbackPackageResponses, emptyPackageResponses, nil
}

func writeFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func getReadmePath(workdir string) (string, error) {
	readmeFilename := "README_GENERATED.md"
	if _, err := os.Stat(filepath.Join(workdir, "README.md")); err != nil {
		if os.IsNotExist(err) {
			readmeFilename = "README.md"
		} else {
			return "", err
		}
	}
	return readmeFilename, nil
}

func fileMapToMd(fileMap map[string]string) string {
	content := ""
	entries := []string{}
	for file := range fileMap {
		entries = append(entries, file)
	}
	slices.Sort(entries)
	for _, file := range entries {
		content += "# " + file + "\n" + fileMap[file] + "\n\n"
	}
	return strings.ReplaceAll(content, "\n", "  \n")
}

func fileMapToString(fileMap map[string]string) string {
	content := ""
	entries := []string{}
	for file := range fileMap {
		entries = append(entries, file)
	}
	slices.Sort(entries)
	for _, file := range entries {
		content += file + "\n" + fileMap[file] + "\n\n"
	}
	return content
}
func getDirFileStructure(workdir string) (string, error) {
	content := fmt.Sprintf("%s directory file structure:\n", filepath.Base(filepath.Dir(workdir)))
	entries := []string{}
	if err := util.WalkDirIgnored(workdir, "", func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(workdir, path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".md") {
			return nil
		}
		entries = append(entries, relPath)
		return nil
	}); err != nil {
		return "", err
	}
	slices.Sort(entries)
	for _, entry := range entries {
		content += "- " + entry + "\n"
	}
	return content, nil
}
