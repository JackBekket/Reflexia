package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Swarmind/libagent/pkg/agent/generic"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/Swarmind/libagent/pkg/tools"
	git "github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	store "github.com/JackBekket/reflexia/pkg"
	packagerunner "github.com/JackBekket/reflexia/pkg/package_runner"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
)

type Config struct {
	GithubLink                    *string
	GithubBranch                  *string
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
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := initConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	workdir, err := processWorkingDirectory(
		*cfg.GithubLink, *cfg.GithubBranch, *cfg.GithubUsername, *cfg.GithubToken)
	if err != nil {
		log.Fatal().Err(err).Msg("process working directory")
	}

	projectConfigVariants, err := project.GetProjectConfig(
		workdir, *cfg.WithConfigFile, cfg.LightCheck,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("get project config")
	}
	projectConfig, err := chooseProjectConfig(projectConfigVariants)
	if err != nil {
		log.Fatal().Err(err).Msg("choose project config")
	}

	agentCfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config.NewConfig")
	}
	if agentCfg.AIURL == "" {
		log.Fatal().Err(err).Msg("empty AI URL")
	}
	if agentCfg.AIToken == "" {
		log.Fatal().Err(err).Msg("empty AI Token")
	}
	if agentCfg.Model == "" {
		log.Fatal().Err(err).Msg("empty model")
	}
	agentCfg.DDGSearchDisable = true
	agentCfg.SemanticSearchDisable = true
	agentCfg.NmapDisable = true
	agentCfg.SimpleCMDExecutorDisable = true
	agentCfg.WebReaderDisable = true
	agentCfg.ReWOODisable = true

	ctx := context.Background()
	agent := generic.Agent{}

	llm, err := openai.New(
		openai.WithBaseURL(agentCfg.AIURL),
		openai.WithToken(agentCfg.AIToken),
		openai.WithModel(agentCfg.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("openai.New")
	}
	agent.LLM = llm

	toolsExecutor, err := tools.NewToolsExecutor(ctx, agentCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("tools.NewToolsExecutor")
	}
	agent.ToolsExecutor = toolsExecutor

	summarizeService := &summarize.SummarizeService{
		Agent: agent,
		LlmOptions: []llms.CallOption{
			llms.WithStopWords(
				projectConfig.StopWords,
			),
			llms.WithRepetitionPenalty(0.7),
		},
		// Tests only
		IgnoreCache:    false,
		OverwriteCache: cfg.OverwriteCache,
		CachePath:      *cfg.CachePath,
	}
	var embeddingsService *store.EmbeddingsService
	if cfg.UseEmbeddings {
		projectName := filepath.Base(projectConfig.RootPath)
		vectorStore, err := store.NewVectorStoreWithPreDelete(
			*cfg.EmbeddingsAIURL,
			*cfg.EmbeddingsAIAPIKey,
			*cfg.EmbeddingsDBURL,
			projectName,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("new vector store")
		}

		embeddingsService = &store.EmbeddingsService{
			Store: vectorStore,
		}
		fmt.Printf("Initialized vector store with %s as project name\n", projectName)
	}

	pkgFiles, err := projectConfig.BuildPackageFiles()
	if err != nil {
		log.Fatal().Err(err).Msg("build package files")
	}

	packageRunnerService := packagerunner.PackageRunnerService{
		PkgFiles:          pkgFiles,
		ProjectConfig:     projectConfig,
		SummarizeService:  summarizeService,
		EmbeddingsService: embeddingsService,
		ExactPackages:     cfg.ExactPackages,
		OverwriteReadme:   cfg.OverwriteReadme,
		WithFileSummary:   cfg.WithFileSummary,
	}

	fallbackFileResponses,
		emptyFileResponses,
		fallbackPackageResponses,
		emptyPackageResponses,
		err := packageRunnerService.RunPackages()
	if err != nil {
		log.Fatal().Err(err).Msg("run packages")
	}

	printEmptyWarning("[WARN] %d fallback attempts for files\n", fallbackFileResponses)
	printEmptyWarning("[WARN] %d empty LLM responses for files\n", emptyFileResponses)
	printEmptyWarning("[WARN] %d fallback attempts for packages\n", fallbackPackageResponses)
	printEmptyWarning("[WARN] %d empty LLM responses for packages\n", emptyPackageResponses)

	if embeddingsService != nil && cfg.EmbeddingsSimSearchTestPrompt != nil {
		results, err := embeddingsService.Store.SimilaritySearch(context.Background(), *cfg.EmbeddingsSimSearchTestPrompt, 2)
		if err != nil {
			log.Fatal().Err(err).Msg("similaity search")
		}
		if len(results) == 0 {
			log.Fatal().Err(err).Msgf("no similarity search results found for %s test prompt", *cfg.EmbeddingsSimSearchTestPrompt)
		}
		fmt.Printf("\n\nSimilarity search results for a test prompt \"%s\":\n", *cfg.EmbeddingsSimSearchTestPrompt)
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
		log.Fatal().Msg("empty environment key")
	}
	return value
}

func chooseProjectConfig(projectConfigVariants map[string]*project.ProjectConfig) (*project.ProjectConfig, error) {
	switch len(projectConfigVariants) {
	case 0:
		return nil, errors.New(
			"failed to detect project language, available languages: go, python, typescript",
		)
	case 1:
		for _, pc := range projectConfigVariants {
			return pc, nil
		}
	default:
		var filenames []string
		for filename := range projectConfigVariants {
			filenames = append(filenames, filename)
		}
		fmt.Println("Multiple project config matches found!")
		for i, filename := range filenames {
			fmt.Printf("%d. %v\n", i+1, filename)
		}
		fmt.Print("Enter the number or filename: ")
		for {
			var input string
			if _, err := fmt.Scanln(&input); err != nil {
				log.Fatal().Err(err).Msg("scanln")
			}
			if index, err := strconv.Atoi(input); err == nil && index > 0 && index <= len(filenames) {
				return projectConfigVariants[filenames[index-1]], nil
			} else {
				for filename, config := range projectConfigVariants {
					if filename == input || strings.TrimSuffix(filename, ".toml") == input {
						return config, nil
					}
				}
			}
		}
	}
	panic("unreachable")
}

func processWorkingDirectory(githubLink, githubBranch, githubUsername, githubToken string) (string, error) {
	workdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if githubLink != "" {
		u, err := url.ParseRequestURI(githubLink)
		if err != nil {
			return "", err
		}

		sPath := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		cdAfter := ""
		if len(sPath) < 2 {
			return "", errors.New("github repository url does not have at least two path elements")
		} else if len(sPath) > 2 {
			// set githubLink to a repository only
			u.Path = strings.Join(sPath[:2], "/")
			githubLink = u.String()

			if sPath[2] != "tree" {
				return "", errors.New("extended github repository url should have 'tree' route after repository name")
			}
			if len(sPath) < 4 {
				return "", errors.New("extended github repository url should have branch name after tree route")
			}
			// do not override branch set from the flag
			if githubBranch == "" {
				githubBranch = sPath[3]
			}
			cdAfter = strings.Join(sPath[4:], "/")
		}

		tempDirEl := []string{workdir, "temp"}
		tempDirEl = append(tempDirEl, sPath[:2]...)

		if githubBranch != "" {
			tempDirEl = append(tempDirEl, githubBranch)
		} else {
			rem := git.NewRemote(memory.NewStorage(), &gitConfig.RemoteConfig{
				Name: "origin",
				URLs: []string{githubLink},
			})
			refs, err := rem.List(&git.ListOptions{})
			if err != nil {
				return "", err
			}

			var defaultBranchRef *plumbing.Reference
			for _, ref := range refs {
				if ref.Name() == "HEAD" {
					defaultBranchRef = ref
					break
				}
			}

			if defaultBranchRef == nil {
				return "", fmt.Errorf("HEAD reference not found in ls-remote output")
			}

			if defaultBranchRef.Type() == plumbing.SymbolicReference {
				targetRefName := defaultBranchRef.Target()
				for _, ref := range refs {
					if ref.Name() == targetRefName {
						defaultBranchRef = ref
						break
					}
				}
			}

			tempDirEl = append(tempDirEl,
				strings.Replace(
					defaultBranchRef.Name().String(),
					"refs/heads/",
					"", 1),
			)
		}

		tempDir := filepath.Join(tempDirEl...)

		workdir = tempDir

		if _, err := os.Stat(workdir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(workdir, os.FileMode(0755)); err != nil {
					return "", err
				}

				cloneOptions := git.CloneOptions{
					URL:               githubLink,
					Depth:             1,
					SingleBranch:      true,
					RecurseSubmodules: git.NoRecurseSubmodules,
					ShallowSubmodules: true,
				}
				if githubBranch != "" {
					cloneOptions.ReferenceName = plumbing.ReferenceName(githubBranch)
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
		workdir = filepath.Join(workdir, cdAfter)

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
		log.Warn().Err(err).Msg("load .env file")
	}

	config := Config{}

	config.GithubLink = flag.String("g", "", "valid link for github repository")
	config.GithubBranch = flag.String("b", "", "optional branch for github repository")
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
