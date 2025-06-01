package reflexia

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/JackBekket/reflexia/internal/github"
	"github.com/JackBekket/reflexia/pkg/config"
	packagerunner "github.com/JackBekket/reflexia/pkg/package_runner"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/store"
	"github.com/JackBekket/reflexia/pkg/summarize"
	"github.com/Swarmind/libagent/pkg/agent/simple"
	agentConfig "github.com/Swarmind/libagent/pkg/config"
	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type ReflexiaCall struct {
	LocalWorkdir     string
	RepositoryURL    string
	RepositoryBranch string
	GithubUsername   string
	GithubToken      string
	WithConfigFile   string
	ExactPackages    string
	CreatePR         bool
	LightCheck       bool
	WithFileSummary  bool
	UseEmbeddings    bool
	OverwriteReadme  bool
	OverwriteCache   bool

	Config      config.Config
	AgentConfig agentConfig.Config
	ChooserFunc func(map[string]*project.ProjectConfig) (*project.ProjectConfig, error)

	PrintTo io.Writer
}

type ReflexiaArtifacts struct {
	PackageRunnerStats packagerunner.RunStats
	PullRequestURL     *string
}

func (o ReflexiaCall) Run(ctx context.Context) (ReflexiaArtifacts, error) {
	artifacts := ReflexiaArtifacts{}

	workdir, err := os.Getwd()
	if err != nil {
		return artifacts, fmt.Errorf("get current workdir: %w", err)
	}

	repo := &git.Repository{}
	branch := ""

	projectName := filepath.Base(workdir)
	cancelFunc := func() {}

	if o.LocalWorkdir != "" {
		workdir = o.LocalWorkdir
		if _, err := os.Stat(workdir); err != nil {
			return artifacts, fmt.Errorf("stat provided workdir: %w", err)
		}
		projectName = filepath.Base(workdir)
	} else if o.RepositoryURL != "" {
		workdir, repo, branch, cancelFunc, err = github.GithubWorkdir(
			workdir,
			o.RepositoryURL,
			o.RepositoryBranch,
			o.GithubUsername,
			o.GithubToken,
		)
		if err != nil {
			cancelFunc()
			return artifacts, fmt.Errorf("prepare github workdir: %w", err)
		}
		projectName = filepath.Base(filepath.Dir(workdir))
	}

	projectConfigVariants, err := project.GetProjectConfig(
		workdir, o.WithConfigFile, o.LightCheck,
	)
	if err != nil {
		cancelFunc()
		return artifacts, fmt.Errorf("get project config: %w", err)
	}
	projectConfig, err := o.ChooserFunc(projectConfigVariants)
	if err != nil {
		cancelFunc()
		return artifacts, fmt.Errorf("choose project config: %w", err)
	}

	agent := &simple.Agent{}

	llm, err := openai.New(
		openai.WithBaseURL(o.AgentConfig.AIURL),
		openai.WithToken(o.AgentConfig.AIToken),
		openai.WithModel(o.AgentConfig.Model),
		openai.WithAPIVersion("v1"),
	)
	if err != nil {
		cancelFunc()
		return artifacts, fmt.Errorf("openai.New: %w", err)
	}
	agent.LLM = llm

	summarizeService := &summarize.SummarizeService{
		Agent: agent,
		LlmOptions: []llms.CallOption{
			llms.WithStopWords(
				projectConfig.StopWords,
			),
			llms.WithRepetitionPenalty(0.7),
		},
		StopWords: projectConfig.StopWords,
		Model:     o.AgentConfig.Model,
		// Tests only
		IgnoreCache:    false,
		OverwriteCache: o.OverwriteCache,
		CachePath:      o.Config.CachePath,
	}
	var embeddingsService *store.EmbeddingsService
	if o.UseEmbeddings {
		vectorStore, err := store.NewVectorStoreWithPreDelete(ctx,
			o.Config.EmbeddingsAIURL,
			o.Config.EmbeddingsAIToken,
			o.Config.EmbeddingsDBURL,
			projectName,
		)
		if err != nil {
			cancelFunc()
			return artifacts, fmt.Errorf("new vector store: %w", err)
		}

		embeddingsService = &store.EmbeddingsService{
			Store: vectorStore,
		}
		fmt.Printf("Initialized vector store with %s as project name\n", projectName)

		if o.Config.EmbeddingsSimSearchTestPrompt != "" {
			defer func() {
				if err := simSearchTest(ctx,
					*embeddingsService,
					o.Config.EmbeddingsSimSearchTestPrompt,
				); err != nil {
					log.Warn().Err(err).Msg("similarity search test")
				}
			}()
		}
	}

	pkgFiles, err := projectConfig.BuildPackageFiles()
	if err != nil {
		cancelFunc()
		return artifacts, fmt.Errorf("build package files: %w", err)
	}

	packageRunnerService := packagerunner.PackageRunnerService{
		PkgFiles:          pkgFiles,
		ProjectConfig:     projectConfig,
		SummarizeService:  summarizeService,
		EmbeddingsService: embeddingsService,
		ExactPackages:     o.ExactPackages,
		OverwriteReadme:   o.OverwriteReadme,
		WithFileSummary:   o.WithFileSummary,

		Model:   o.AgentConfig.Model,
		PrintTo: o.PrintTo,
	}

	artifacts.PackageRunnerStats, err = packageRunnerService.RunPackages(ctx)
	if err != nil {
		cancelFunc()
		return artifacts, fmt.Errorf("run packages: %w", err)
	}

	if o.CreatePR {
		prURL, err := github.CreatePR(ctx,
			repo,
			branch,
			o.GithubUsername,
			o.GithubToken,
		)
		if err != nil {
			cancelFunc()
			return artifacts, fmt.Errorf("creating pull request: %w", err)
		}

		artifacts.PullRequestURL = &prURL
	}

	cancelFunc()
	return artifacts, nil
}

func simSearchTest(ctx context.Context, embeddingsService store.EmbeddingsService, testPrompt string) error {
	results, err := embeddingsService.Store.SimilaritySearch(ctx, testPrompt, 2)
	if err != nil {
		return fmt.Errorf("similarity search: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no similarity search results found for %s test prompt", testPrompt)
	}
	fmt.Printf("\n\nSimilarity search results for a test prompt \"%s\":\n", testPrompt)
	for i, result := range results {
		fmt.Printf("%d: score %f\n", i, result.Score)
		for k, v := range result.Metadata {
			fmt.Printf("    %s: %s\n", k, v)
		}
		fmt.Printf("\n%s\n\n", result.PageContent)
	}

	return nil
}
