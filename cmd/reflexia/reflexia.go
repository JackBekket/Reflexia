package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"

	util "github.com/JackBekket/reflexia/internal"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
)

type Config struct {
	GithubLink      *string
	GithubUsername  *string
	GithubToken     *string
	WithConfigFile  *string
	ExactPackages   *string
	LightCheck      bool
	WithFileSummary bool
	OverwriteReadme bool
	OverwriteCache  bool
	CachePath       *string
}

func main() {
	config, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}

	workdir, err := processWorkingDirectory(
		*config.GithubLink, *config.GithubUsername, *config.GithubToken)
	if err != nil {
		log.Fatal(err)
	}

	projectConfig, err := project.GetProjectConfig(
		workdir, *config.WithConfigFile, config.LightCheck,
	)
	if err != nil {
		log.Fatal(err)
	}
	summarizeService := &summarize.SummarizerService{
		HelperURL: loadEnv("HELPER_URL"),
		Model:     loadEnv("MODEL"),
		ApiToken:  loadEnv("API_TOKEN"),
		Network:   "local",
		LlmOptions: []llms.CallOption{
			llms.WithStopWords(
				projectConfig.StopWords,
			),
			llms.WithRepetitionPenalty(0.7),
		},
		OverwriteCache: config.OverwriteCache,
		CachePath:      *config.CachePath,
	}

	// Generates summary for a file content
	// return the map of fileMap[relPath] = generation_content
	fileMap, err := summarizeService.SummarizeCode(projectConfig)
	if err != nil {
		log.Fatal(err)
	}

	pkgFiles, err := projectConfig.BuildPackageFiles()
	if err != nil {
		log.Fatal(err)
	}

	for pkg, files := range pkgFiles {
		if *config.ExactPackages != "" &&
			!slices.Contains(strings.Split(*config.ExactPackages, ","), pkg) {
			continue
		}
		fmt.Printf("\nPackage %s\n", pkg)

		pkgFileMap := map[string]string{}
		pkgDir := ""
		for file, content := range fileMap {
			if slices.Contains(files, file) {
				if pkgDir == "" {
					pkgDir = filepath.Join(workdir, filepath.Dir(file))
				}
				pkgFileMap[file] = content
			}
		}
		if pkgDir == "" {
			log.Println("There is no mathcing files for pkg: ", pkg)
			continue
		}

		projectStructure, err := getDirFileStructure(pkgDir)
		if err != nil {
			log.Print(err)
		}
		fmt.Println(projectStructure)

		// Generate Summary for a package (summarizing and group file summarization by package name)
		// Get whole map of code summaries as a string and toss it to summarize .MD for a package
		pkgSummaryContent, err := summarizeService.SummarizeRequest(
			projectConfig.PackagePrompt,
			fileMapToString(pkgFileMap),
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\nSummary for a package %s: \n", pkg)
		fmt.Println(pkgSummaryContent)

		readmeFilename := "README.md"
		if !config.OverwriteReadme {
			readmeFilename, err = getReadmePath(pkgDir)
			if err != nil {
				log.Fatal(err)
			}
		}
		if err := writeFile(
			filepath.Join(pkgDir, readmeFilename),
			pkgSummaryContent,
		); err != nil {
			log.Fatal(err)
		}

		if config.WithFileSummary {
			if err := writeFile(
				filepath.Join(pkgDir, "FILES.md"),
				fileMapToMd(pkgFileMap),
			); err != nil {
				log.Fatal(err)
			}
		}

	}
}

func loadEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("empty environment key %s", key)
	}
	return value
}

func getDirFileStructure(workdir string) (string, error) {
	content := fmt.Sprintf("%s directory file structure:\n", filepath.Base(workdir))
	if err := util.WalkDirIgnored(workdir, "", func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(workdir, path)
		if err != nil {
			return err
		}
		content += "- " + relPath + "\n"
		return nil
	}); err != nil {
		return "", err
	}
	return content, nil
}

func fileMapToString(fileMap map[string]string) string {
	content := ""
	for file, summary := range fileMap {
		content += file + "\n" + summary + "\n\n"
	}
	return content
}

func fileMapToMd(fileMap map[string]string) string {
	content := ""
	for file, summary := range fileMap {
		content += "# " + file + "\n" + summary + "\n\n"
	}
	return strings.ReplaceAll(content, "\n", "  \n")
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

func processWorkingDirectory(githubLink, githubUsername, githubToken string) (string, error) {
	workdir := loadEnv("PWD")

	if githubLink != "" {
		u, err := url.ParseRequestURI(githubLink)
		if err != nil {
			return "", err
		}

		sPath := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(sPath) != 2 {
			return "", errors.New("github repository url does not have two path elements")
		}

		tempDirEl := []string{workdir, "temp"}
		tempDirEl = append(tempDirEl, sPath...)
		tempDir := filepath.Join(tempDirEl...)

		workdir = tempDir

		if _, err := os.Stat(workdir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(workdir, os.FileMode(0755)); err != nil {
					return "", err
				}

				cloneOptions := git.CloneOptions{
					URL:               githubLink,
					RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
					Depth:             1,
				}
				if githubUsername != "" && githubToken != "" {
					cloneOptions.Auth = &http.BasicAuth{
						Username: githubUsername,
						Password: githubToken,
					}
				}

				if _, err := git.PlainClone(workdir, false, &cloneOptions); err != nil {
					if err := os.RemoveAll(workdir); err != nil {
						return "", err
					}
					return "", err
				}

			} else {
				return "", err
			}
		}
	} else if len(flag.Args()) > 0 {
		workdir = flag.Arg(0)
		if _, err := os.Stat(workdir); err != nil {
			return "", err
		}
	}

	return workdir, nil
}

func initConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	config := Config{}

	config.GithubLink = flag.String("g", "", "valid link for github repository")
	config.GithubUsername = flag.String("u", "", "github username for ssh auth")
	config.ExactPackages = flag.String("p", "", "exact package names, ',' delimited")

	githubToken := os.Getenv("GH_TOKEN")
	config.GithubToken = &githubToken
	flag.StringVar(config.GithubToken, "t", *config.GithubToken, "github token for ssh auth")

	config.WithConfigFile = flag.String("l", "", "config filename in project_config to use")

	cachePath := os.Getenv("CACHE_PATH")
	config.CachePath = &cachePath
	flag.StringVar(config.CachePath, "a", *config.CachePath, "cache folder path (defaults to .reflexia_cache)")
	if *config.CachePath == "" {
		cachePath = ".reflexia_cache"
		config.CachePath = &cachePath
	}

	config.LightCheck = false
	config.WithFileSummary = false
	config.OverwriteReadme = false
	config.OverwriteCache = false

	flag.BoolFunc("c",
		"do not check project root folder specific files such as go.mod or package.json",
		func(_ string) error {
			config.LightCheck = true
			return nil
		})
	flag.BoolFunc("f",
		"Save individual file summary intermediate result to the FILES.md",
		func(_ string) error {
			config.WithFileSummary = true
			return nil
		})
	flag.BoolFunc("r",
		"overwrite README.md instead of README_GENERATED.md creation/overwrite",
		func(_ string) error {
			config.OverwriteReadme = true
			return nil
		})
	flag.BoolFunc("d",
		"Overwrite generated summary caches",
		func(_ string) error {
			config.OverwriteCache = true
			return nil
		})

	flag.Parse()

	return &config, nil
}
