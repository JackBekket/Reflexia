package cli

import (
	"strings"
	"testing"

	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
	"github.com/Swarmind/libagent/pkg/agent/simple"
	"github.com/Swarmind/libagent/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestProcessWorkingDirectory(t *testing.T) {
	t.Run(
		"Working directory without links",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("", "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(workdir, "Reflexia") {
				t.Fatalf("workdir %s is not Reflexia", workdir)
			}
		},
	)
	t.Run(
		"Working directory with github link",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("https://github.com/JackBekket/GitHelper", "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(workdir, "GitHelper") {
				t.Fatalf("workdir %s is not GitHelper", workdir)
			}
		},
	)
	t.Run(
		"Working directory with github link using cached temp result",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("https://github.com/JackBekket/GitHelper", "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(workdir, "GitHelper") {
				t.Fatalf("workdir %s is not GitHelper", workdir)
			}
		},
	)
}

func TestProcessWorkingDirectoryIgnore(t *testing.T) {
	t.Run(
		"Working directory with github link",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("https://github.com/JackBekket/LocalAI", "", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(workdir, "LocalAI") {
				t.Fatalf("workdir %s is not LocalAI", workdir)
			}
		},
	)

}

func TestGetProjectConfig(t *testing.T) {
	t.Run(
		"Project language .toml config",
		func(t *testing.T) {
			projectConfigVariants, err := project.GetProjectConfig(
				"temp/JackBekket/GitHelper/master", "", false,
			)
			if err != nil {
				t.Fatal(err)
			}
			failed := true
			for filename := range projectConfigVariants {
				if filename == "go.toml" {
					failed = false
				}
			}
			if failed {
				t.Fatal("no go.toml in go project detected!")
			}
		},
	)
}

func TestBuildPackageFiles(t *testing.T) {
	t.Run(
		"Project package files",
		func(t *testing.T) {
			projectConfigVariants, err := project.GetProjectConfig(
				"temp/JackBekket/GitHelper/master", "", false,
			)
			if err != nil {
				t.Fatal(err)
			}
			var projectConfig *project.ProjectConfig
			for filename, config := range projectConfigVariants {
				if filename == "go.toml" {
					projectConfig = config
					break
				}
			}
			pkgFiles, err := projectConfig.BuildPackageFiles()
			if err != nil {
				t.Fatal(err)
			}
			failed := true
			for pkg := range pkgFiles {
				if pkg == "main" {
					failed = false
					break
				}
			}
			if failed {
				t.Fatal("no main package in JackBekket/GitHelper project")
			}
		},
	)
}

func TestLLMResponse(t *testing.T) {
	projectConfigVariants, err := project.GetProjectConfig(
		"temp/JackBekket/GitHelper/master", "", false,
	)
	if err != nil {
		t.Fatal(err)
	}
	var projectConfig *project.ProjectConfig
	for filename, config := range projectConfigVariants {
		if filename == "go.toml" {
			projectConfig = config
			break
		}
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

	agent := &simple.Agent{}

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

	summarizeService := &summarize.SummarizeService{
		Agent: agent,
		LlmOptions: []llms.CallOption{
			llms.WithStopWords(
				projectConfig.StopWords,
			),
			llms.WithRepetitionPenalty(0.7),
		},
		IgnoreCache: true,
		StopWords:   projectConfig.StopWords,
		Model:       agentCfg.Model,
	}
	t.Run(
		"Test LLM response",
		func(t *testing.T) {
			testPrompt := "This is a test prompt. Write a hex notation of red color in reply."
			resp, err := summarizeService.LLMRequest(testPrompt)
			if err != nil {
				t.Fatal(err)
			}
			if resp == "" {
				t.Fatal("Empty output from LLM")
			}
		},
	)
}
