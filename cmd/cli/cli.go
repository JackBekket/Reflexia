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

	flag.StringVar(&cfg.CachePath, "a", cfg.CachePath, "cache folder path (defaults to .reflexia_cache)")
	flag.StringVar(&cfg.EmbeddingsAIURL, "eu", cfg.EmbeddingsAIURL, "embeddings AI URL")
	flag.StringVar(&cfg.EmbeddingsAIToken, "ea", cfg.EmbeddingsAIToken, "embeddings AI API Key")
	flag.StringVar(&cfg.EmbeddingsDBURL, "ed", cfg.EmbeddingsDBURL, "embeddings pgxpool DB connect URL")
	flag.StringVar(&cfg.EmbeddingsSimSearchTestPrompt, "et", cfg.EmbeddingsSimSearchTestPrompt, "embeddings similarity search validation test prompt")

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
	flag.StringVar(&reflexiaOpts.RepositoryURL, "g", "", "valid link for github repository")
	flag.StringVar(&reflexiaOpts.RepositoryBranch, "b", "", "optional branch for github repository")
	flag.StringVar(&reflexiaOpts.GithubUsername, "u", "", "github username for ssh auth")
	flag.StringVar(&reflexiaOpts.GithubToken, "t", reflexiaOpts.GithubToken, "github token for ssh auth")

	flag.StringVar(&reflexiaOpts.WithConfigFile, "l", "", "config filename in project_config to use")
	flag.StringVar(&reflexiaOpts.ExactPackages, "p", "", "exact package names, ',' delimited")

	flag.BoolFunc("w",
		"do a commit to a _autodoc suffixed branch and raise a PR",
		func(_ string) error {
			reflexiaOpts.CreatePR = true
			return nil
		})
	flag.BoolFunc("c",
		"do not check project root folder specific files such as go.mod or package.json",
		func(_ string) error {
			reflexiaOpts.LightCheck = true
			return nil
		})
	flag.BoolFunc("f",
		"save individual file summary intermediate result to the FILES.md",
		func(_ string) error {
			reflexiaOpts.WithFileSummary = true
			return nil
		})
	flag.BoolFunc("r",
		"overwrite README.md instead of README_GENERATED.md creation/overwrite",
		func(_ string) error {
			reflexiaOpts.OverwriteReadme = true
			return nil
		})
	flag.BoolFunc("d",
		"overwrite generated summary caches",
		func(_ string) error {
			reflexiaOpts.OverwriteCache = true
			return nil
		})
	flag.BoolFunc("e", "Use Embeddings",
		func(_ string) error {
			reflexiaOpts.UseEmbeddings = true
			return nil
		})

	flag.Parse()

	return reflexiaOpts, nil
}
