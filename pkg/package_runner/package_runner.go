package packagerunner

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/JackBekket/reflexia/internal/util"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/store"
	"github.com/JackBekket/reflexia/pkg/summarize"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/schema"
)

type RunStats struct {
	FallbackFileResponses    []string
	EmptyFileResponses       []string
	FallbackPackageResponses []string
	EmptyPackageResponses    []string
}

type PackageRunnerService struct {
	PkgFiles          map[string][]string
	ProjectConfig     *project.ProjectConfig
	SummarizeService  *summarize.SummarizeService
	EmbeddingsService *store.EmbeddingsService
	ExactPackages     string
	OverwriteReadme   bool
	WithFileSummary   bool

	Model   string
	PrintTo io.Writer
}

func (s *PackageRunnerService) RunPackages(ctx context.Context) (RunStats, error) {
	stats := RunStats{}

	pcPrompts, ok := s.ProjectConfig.Prompts["default"]
	if !ok {
		return stats, fmt.Errorf("failed to load default prompts")
	}

	for model, prompts := range s.ProjectConfig.Prompts {
		if model == "default" {
			continue
		}

		matched, err := regexp.MatchString(model, s.Model)
		if err == nil && matched {
			pcPrompts = prompts
			break
		}
	}

	codePrompt := pcPrompts.CodePrompt
	codePromptFallback := ""
	if pcPrompts.CodePromptFallback != nil {
		codePromptFallback = *pcPrompts.CodePromptFallback
	}
	packagePrompt := pcPrompts.PackagePrompt
	packagePromptFallback := ""
	if pcPrompts.PackagePromptFallback != nil {
		packagePromptFallback = *pcPrompts.PackagePromptFallback
	}

	for pkg, files := range s.PkgFiles {
		if s.ExactPackages != "" &&
			!slices.Contains(strings.Split(s.ExactPackages, ","), pkg) {
			continue
		}
		fmt.Fprintf(s.PrintTo, "Package %s\n", pkg)

		pkgFileMap := map[string]string{}
		pkgDir := ""
		for _, relPath := range files {
			fmt.Fprintf(s.PrintTo, "%s\n", relPath)
			content, err := os.ReadFile(filepath.Join(s.ProjectConfig.RootPath, relPath))
			if err != nil {
				return stats, err
			}

			contentStr := string(content)
			codeSummaryContent := "Empty file"

			if strings.TrimSpace(contentStr) != "" {
				codeSummaryContent, err = s.SummarizeService.LLMRequest(ctx,
					"%s```\n%s\n```",
					codePrompt, contentStr,
				)
				if err != nil {
					return stats, err
				}
				if strings.TrimSpace(codeSummaryContent) == "" &&
					codePromptFallback != "" {
					stats.FallbackFileResponses = append(
						stats.FallbackFileResponses,
						filepath.Join(s.ProjectConfig.RootPath, relPath),
					)
					codeSummaryContent, err = s.SummarizeService.LLMRequest(ctx,
						"%s```\n%s\n```",
						codePromptFallback, contentStr,
					)
					if err != nil {
						return stats, err
					}
				}
			}
			if strings.TrimSpace(codeSummaryContent) == "" {
				stats.EmptyFileResponses = append(
					stats.EmptyFileResponses,
					filepath.Join(s.ProjectConfig.RootPath, relPath),
				)
				fmt.Fprintf(s.PrintTo, "[WARN] empty file summary!\n")
			} else {
				fmt.Fprintf(s.PrintTo, "%s\n", codeSummaryContent)
			}
			fmt.Fprintf(s.PrintTo, "\n")

			pkgFileMap[relPath] = codeSummaryContent
			if s.EmbeddingsService != nil {
				ids, err := s.EmbeddingsService.Store.AddDocuments(ctx,
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
					return stats, err
				}
				fmt.Fprintf(s.PrintTo, "Succesfully pushed docs %s into embeddings vector store\n", ids)
			}

			if pkgDir == "" {
				pkgDir = filepath.Join(s.ProjectConfig.RootPath, filepath.Dir(relPath))
			}
		}
		if pkgDir == "" {
			log.Info().Msgf("There is no mathcing files for pkg: %s", pkg)
			continue
		}

		fileStructure, err := getDirFileStructure(pkg, pkgDir)
		if err != nil {
			log.Warn().Err(err).Msg("getDirFileStructure error")
		}

		fmt.Fprintf(s.PrintTo, "Summary for a package %s: \n", pkg)
		// Generate Summary for a package (summarizing and group file summarization by package name)
		// Get whole map of code summaries as a string and toss it to summarize .MD for a package
		pkgSummaryContent, err := s.SummarizeService.LLMRequest(ctx,
			"%s\n\n%s\n%s",
			packagePrompt,
			fileStructure,
			fileMapToString(pkgFileMap),
		)
		if err != nil {
			return stats, err
		}

		if strings.TrimSpace(pkgSummaryContent) == "" && packagePromptFallback != "" {
			stats.FallbackPackageResponses = append(stats.FallbackPackageResponses, pkg)
			pkgSummaryContent, err = s.SummarizeService.LLMRequest(ctx,
				"%s\n\n%s\n%s",
				packagePromptFallback,
				fileStructure,
				fileMapToString(pkgFileMap),
			)
			if err != nil {
				return stats, err
			}
		}

		if strings.TrimSpace(pkgSummaryContent) == "" {
			stats.EmptyPackageResponses = append(stats.EmptyPackageResponses, pkg)
			fmt.Fprintf(s.PrintTo, "[WARN] empty package summary\n")
		} else {
			fmt.Fprintf(s.PrintTo, "%s\n", pkgSummaryContent)
		}
		fmt.Fprintf(s.PrintTo, "\n")

		if s.EmbeddingsService != nil {
			ids, err := s.EmbeddingsService.Store.AddDocuments(ctx,
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
				return stats, err
			}
			fmt.Fprintf(s.PrintTo, "Succesfully pushed docs %s into embeddings vector store\n", ids)
		}

		readmeFilename := "README.md"
		if !s.OverwriteReadme {
			readmeFilename, err = getReadmePath(pkgDir)
			if err != nil {
				return stats, err
			}
		}
		if err := writeFile(
			filepath.Join(pkgDir, readmeFilename),
			pkgSummaryContent,
		); err != nil {
			return stats, err
		}

		if s.WithFileSummary {
			if err := writeFile(
				filepath.Join(pkgDir, "FILES.md"),
				fileMapToMd(pkgFileMap),
			); err != nil {
				return stats, err
			}
		}
	}
	return stats, nil
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

func getDirFileStructure(pkg, workdir string) (string, error) {
	content := fmt.Sprintf("%s directory file structure:\n", pkg)
	entries := []string{}
	if err := util.WalkDirIgnored(
		workdir, filepath.Join(workdir, ".gitignore"),
		func(path string, d fs.DirEntry) error {
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
		},
	); err != nil {
		return "", err
	}
	slices.Sort(entries)
	for _, entry := range entries {
		content += "- " + entry + "\n"
	}
	return content, nil
}
