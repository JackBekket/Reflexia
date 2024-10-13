package main

import (
	"context"
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
	"github.com/tmc/langchaingo/schema"

	util "github.com/JackBekket/reflexia/internal"
	store "github.com/JackBekket/reflexia/pkg"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
)

type Config struct {
	GithubLink                    *string
	GithubUsername                *string
	GithubToken                   *string
	WithConfigFile                *string
	ExactPackages                 *string
	LightCheck                    bool
	WithFileSummary               bool
	UseEmbeddings                 bool
	OverwriteReadme               bool
	OverwriteCache                bool
	EmbeddingsAIURL               *string
	EmbeddingsAIAPIKey            *string
	EmbeddingsDBURL               *string
	EmbeddingsSimSearchTestPrompt *string
	CachePath                     *string
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
	var embeddingsService *store.EmbeddingsService
	if config.UseEmbeddings {
		projectName := filepath.Base(projectConfig.RootPath)
		vectorStore, err := store.NewVectorStoreWithPreDelete(
			*config.EmbeddingsAIURL,
			*config.EmbeddingsAIAPIKey,
			*config.EmbeddingsDBURL,
			projectName,
		)
		if err != nil {
			log.Fatal(err)
		}

		embeddingsService = &store.EmbeddingsService{
			Store: vectorStore,
		}
		fmt.Printf("Initialized vector store with %s as project name\n", projectName)
	}

	pkgFiles, err := projectConfig.BuildPackageFiles()
	if err != nil {
		log.Fatal(err)
	}

	fallbackFileResponses := []string{}
	emptyFileResponses := []string{}
	fallbackPackageResponses := []string{}
	emptyPackageResponses := []string{}

	for pkg, files := range pkgFiles {
		if *config.ExactPackages != "" &&
			!slices.Contains(strings.Split(*config.ExactPackages, ","), pkg) {
			continue
		}
		fmt.Printf("\nPackage %s\n", pkg)

		pkgFileMap := map[string]string{}
		pkgDir := ""
		for _, relPath := range files {
			fmt.Println(relPath)
			content, err := os.ReadFile(filepath.Join(projectConfig.RootPath, relPath))
			if err != nil {
				log.Fatal(err)
			}

			contentStr := string(content)
			codeSummaryContent := "Empty file"

			if strings.TrimSpace(contentStr) != "" {
				codeSummaryContent, err = summarizeService.LLMRequest(
					"%s```%s```",
					projectConfig.CodePrompt, contentStr,
				)
				if err != nil {
					log.Fatal(err)
				}
				if strings.TrimSpace(codeSummaryContent) == "" &&
					projectConfig.CodePromptFallback != nil {
					fallbackFileResponses = append(
						fallbackFileResponses,
						filepath.Join(projectConfig.RootPath, relPath),
					)
					codeSummaryContent, err = summarizeService.LLMRequest(
						"%s```%s```",
						*projectConfig.CodePromptFallback, contentStr,
					)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				fmt.Println("Empty file")
			}

			fmt.Printf("\n")
			pkgFileMap[relPath] = codeSummaryContent
			if embeddingsService != nil {
				ids, err := embeddingsService.Store.AddDocuments(
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
					log.Fatal(err)
				}
				fmt.Printf("Succesfully pushed docs %s into embeddings vector store\n", ids)
			}

			if strings.TrimSpace(codeSummaryContent) == "" {
				emptyFileResponses = append(
					emptyFileResponses,
					filepath.Join(projectConfig.RootPath, relPath),
				)
			}
			if pkgDir == "" {
				pkgDir = filepath.Join(projectConfig.RootPath, filepath.Dir(relPath))
			}
		}
		if pkgDir == "" {
			log.Println("There is no mathcing files for pkg: ", pkg)
			continue
		}

		fileStructure, err := getDirFileStructure(pkgDir)
		if err != nil {
			log.Print(err)
		}

		fmt.Printf("\nSummary for a package %s: \n", pkg)
		// Generate Summary for a package (summarizing and group file summarization by package name)
		// Get whole map of code summaries as a string and toss it to summarize .MD for a package
		pkgSummaryContent, err := summarizeService.LLMRequest(
			"%s\n\n%s\n%s",
			projectConfig.PackagePrompt,
			fileStructure,
			fileMapToString(pkgFileMap),
		)
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(pkgSummaryContent) == "" && projectConfig.PackagePromptFallback != nil {
			fallbackPackageResponses = append(fallbackPackageResponses, pkg)
			pkgSummaryContent, err = summarizeService.LLMRequest(
				"%s\n\n%s\n%s",
				*projectConfig.PackagePromptFallback,
				fileStructure,
				fileMapToString(pkgFileMap),
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if strings.TrimSpace(pkgSummaryContent) == "" {
			emptyPackageResponses = append(emptyPackageResponses, pkg)
		}

		if embeddingsService != nil {
			ids, err := embeddingsService.Store.AddDocuments(
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
				log.Fatal(err)
			}
			fmt.Printf("Succesfully pushed docs %s into embeddings vector store\n", ids)
		}

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

	printEmptyWarning("[WARN] %d fallback attempts for files\n", fallbackFileResponses)
	printEmptyWarning("[WARN] %d empty LLM responses for files\n", emptyFileResponses)
	printEmptyWarning("[WARN] %d fallback attempts for packages\n", fallbackPackageResponses)
	printEmptyWarning("[WARN] %d empty LLM responses for packages\n", emptyPackageResponses)

	if embeddingsService != nil && config.EmbeddingsSimSearchTestPrompt != nil {
		results, err := embeddingsService.Store.SimilaritySearch(context.Background(), *config.EmbeddingsSimSearchTestPrompt, 2)
		if err != nil {
			log.Fatal(err)
		}
		if len(results) == 0 {
			log.Fatalf("No similarity search results found for %s test prompt\n", *config.EmbeddingsSimSearchTestPrompt)
		}
		fmt.Printf("\n\nSimilarity search results for a test prompt \"%s\":\n", *config.EmbeddingsSimSearchTestPrompt)
		for i, result := range results {
			fmt.Printf("%d: score %f\n", i, result.Score)
			for k, v := range result.Metadata {
				fmt.Printf("    %s: %s\n", k, v)
			}
			fmt.Printf("\n%s\n\n", result.PageContent)
		}
	}
}

func printEmptyWarning(message string, responses []string) {
	if len(responses) == 0 {
		return
	}
	fmt.Printf(message, len(responses))
	for _, response := range responses {
		fmt.Printf(" - %s\n", response)
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
		log.Println(err)
	}

	config := Config{}

	config.GithubLink = flag.String("g", "", "valid link for github repository")
	config.GithubUsername = flag.String("u", "", "github username for ssh auth")

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

	embAIURL := os.Getenv("EMBEDDINGS_AI_URL")
	config.EmbeddingsAIURL = &embAIURL
	flag.StringVar(config.EmbeddingsAIURL, "eu", *config.EmbeddingsAIURL, "Embeddings AI URL")

	embAIAPIKey := os.Getenv("EMBEDDINGS_AI_KEY")
	config.EmbeddingsAIAPIKey = &embAIAPIKey
	flag.StringVar(config.EmbeddingsAIAPIKey, "ea", *config.EmbeddingsAIAPIKey, "Embeddings AI API Key")

	embDBURL := os.Getenv("EMBEDDINGS_DB_URL")
	config.EmbeddingsDBURL = &embDBURL
	flag.StringVar(config.EmbeddingsDBURL, "ed", *config.EmbeddingsDBURL, "Embeddings pgxpool DB connect URL")

	embSimSearchTestPrompt := os.Getenv("EMBEDDINGS_SIM_SEARCH_TEST_PROMPT")
	config.EmbeddingsSimSearchTestPrompt = &embSimSearchTestPrompt
	flag.StringVar(config.EmbeddingsSimSearchTestPrompt, "et", *config.EmbeddingsSimSearchTestPrompt, "Embeddings similarity search validation test prompt")

	config.ExactPackages = flag.String("p", "", "exact package names, ',' delimited")
	config.LightCheck = false
	config.WithFileSummary = false
	config.OverwriteReadme = false
	config.OverwriteCache = false
	config.UseEmbeddings = false

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
	flag.BoolFunc("e", "Use Embeddings",
		func(_ string) error {
			config.UseEmbeddings = true
			return nil
		})

	flag.Parse()

	return &config, nil
}
