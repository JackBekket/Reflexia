package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/JackBekket/reflexia/pkg/config"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/reflexia"

	agentConfig "github.com/Swarmind/libagent/pkg/config"
)

func Run(cfg config.Config) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx := context.Background()

	reflexiaOpts, err := readReflexiaCall(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	if reflexiaOpts.RepositoryURL == "" && len(flag.Args()) > 0 {
		wd := flag.Arg(0)
		reflexiaOpts.LocalWorkdir = wd
	}

	artifacts, err := reflexiaOpts.Run(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("reflexia run")
	}

	printEmptyWarning(
		"[WARN] %d fallback attempts for files\n",
		artifacts.PackageRunnerStats.FallbackFileResponses,
	)
	printEmptyWarning(
		"[WARN] %d empty LLM responses for files\n",
		artifacts.PackageRunnerStats.EmptyFileResponses,
	)
	printEmptyWarning(
		"[WARN] %d fallback attempts for packages\n",
		artifacts.PackageRunnerStats.FallbackPackageResponses,
	)
	printEmptyWarning(
		"[WARN] %d empty LLM responses for packages\n",
		artifacts.PackageRunnerStats.EmptyPackageResponses,
	)

	if artifacts.PullRequestURL != nil {
		fmt.Printf("Pull request created: %s\n", *artifacts.PullRequestURL)
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

func readReflexiaCall(cfg config.Config) (reflexia.ReflexiaCall, error) {
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("load .env file")
	}

	flag.StringVar(&cfg.CachePath, "cache-path", cfg.CachePath, "cache folder path (defaults to .reflexia_cache)")
	flag.StringVar(&cfg.EmbeddingsAIURL, "embeddings-url", cfg.EmbeddingsAIURL, "embeddings AI URL")
	flag.StringVar(&cfg.EmbeddingsAIToken, "embeddings-token", cfg.EmbeddingsAIToken, "embeddings AI API Key")
	flag.StringVar(&cfg.EmbeddingsDBURL, "embeddings-db", cfg.EmbeddingsDBURL, "embeddings pgxpool DB connect URL")
	flag.StringVar(&cfg.EmbeddingsSimSearchTestPrompt, "embeddings-test-prompt", cfg.EmbeddingsSimSearchTestPrompt, "embeddings similarity search validation test prompt")

	agentCfg, err := agentConfig.NewConfig()
	if err != nil {
		return reflexia.ReflexiaCall{}, fmt.Errorf("config.NewConfig: %w", err)
	}

	reflexiaOpts := reflexia.ReflexiaCall{
		Config:      cfg,
		AgentConfig: agentCfg,
		ChooserFunc: project.CLIChooser,
		PrintTo:     os.Stdout,
	}

	reflexiaOpts.GithubToken = os.Getenv("GH_TOKEN")
	flag.StringVar(&reflexiaOpts.RepositoryURL, "gh-url", "", "valid link for github repository")
	flag.StringVar(&reflexiaOpts.RepositoryBranch, "branch", "", "optional branch for github repository")
	flag.StringVar(&reflexiaOpts.GithubUsername, "gh-username", "", "github username for ssh auth")
	flag.StringVar(&reflexiaOpts.GithubToken, "gh-token", reflexiaOpts.GithubToken, "github token for ssh auth")

	flag.StringVar(&reflexiaOpts.WithConfigFile, "cfg", "", "config filename in project_config to use")
	flag.StringVar(&reflexiaOpts.ExactPackages, "pkgs", "", "exact package names, ',' delimited")

	flag.BoolFunc("gh-preserve-workdir",
		"do not remove github repository workdir (only for CLI mode)",
		func(_ string) error {
			reflexiaOpts.PreserveGithubWorkdir = true
			return nil
		})
	flag.BoolFunc("pr",
		"do a commit to a _autodoc suffixed branch and raise a PR",
		func(_ string) error {
			reflexiaOpts.CreatePR = true
			return nil
		})
	flag.BoolFunc("light-check",
		"do not check project root folder specific files such as go.mod or package.json",
		func(_ string) error {
			reflexiaOpts.LightCheck = true
			return nil
		})
	flag.BoolFunc("use-files",
		"save individual file summary intermediate result to the FILES.md",
		func(_ string) error {
			reflexiaOpts.WithFileSummary = true
			return nil
		})
	flag.BoolFunc("ow-readme",
		"overwrite README.md instead of README_GENERATED.md creation/overwrite",
		func(_ string) error {
			reflexiaOpts.OverwriteReadme = true
			return nil
		})
	flag.BoolFunc("ow-cache",
		"overwrite generated summary caches",
		func(_ string) error {
			reflexiaOpts.OverwriteCache = true
			return nil
		})
	flag.BoolFunc("dry-run",
		"only check if project language is determined",
		func(_ string) error {
			reflexiaOpts.DryRun = true
			return nil
		})
	flag.BoolFunc("embeddings", "Use Embeddings",
		func(_ string) error {
			reflexiaOpts.UseEmbeddings = true
			return nil
		})

	flag.Parse()

	return reflexiaOpts, nil
}
