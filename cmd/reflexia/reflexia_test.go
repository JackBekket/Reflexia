package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/summarize"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
)

func TestProcessWorkingDirectory(t *testing.T) {
	// since test is running in cmd/reflexia folder we need to go back
	err := os.Chdir("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"Working directory without links",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(workdir) != "Reflexia" {
				t.Fatalf("workdir != reflexia")
			}
		},
	)
	t.Run(
		"Working directory with github link",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("https://github.com/JackBekket/GitHelper", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(workdir) != "GitHelper" {
				t.Fatalf("workdir != GitHelper")
			}
		},
	)
	t.Run(
		"Working directory with github link using cached temp result",
		func(t *testing.T) {
			workdir, err := processWorkingDirectory("https://github.com/JackBekket/GitHelper", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(workdir) != "GitHelper" {
				t.Fatalf("workdir != GitHelper")
			}
		},
	)
}

func TestGetProjectConfig(t *testing.T) {
	t.Run(
		"Project language .toml config",
		func(t *testing.T) {
			projectConfigVariants, err := project.GetProjectConfig(
				"temp/JackBekket/GitHelper", "", false,
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
				"temp/JackBekket/GitHelper", "", false,
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
	helper_url, model, apiToken := checkLoadLLMConfig(t)

	projectConfigVariants, err := project.GetProjectConfig(
		"temp/JackBekket/GitHelper", "", false,
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

	summarizeService := &summarize.SummarizerService{
		HelperURL: helper_url,
		Model:     model,
		ApiToken:  apiToken,
		Network:   "local",
		LlmOptions: []llms.CallOption{
			llms.WithStopWords(
				projectConfig.StopWords,
			),
			llms.WithRepetitionPenalty(0.7),
		},
		IgnoreCache: true,
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

func checkLoadLLMConfig(t *testing.T) (string, string, string) {
	if err := godotenv.Load(); err != nil {
		t.Skip("Failed to load .env file. LLM tests will be skipped")
	}
	helper_url := os.Getenv("HELPER_URL")
	if helper_url == "" {
		t.Skip("No HELPER_URL environment variable found. LLM tests will be skipped")
	}
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Get(helper_url)
	if err != nil {
		t.Skip("HELPER_URL is unavailable. LLM tests will be skipped")
	}
	model := os.Getenv("MODEL")
	if model == "" {
		t.Skip("No MODEL environment variable found. LLM tests will be skipped")
	}
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		t.Skip("No API_TOKEN environment variable found. LLM tests will be skipped")
	}
	return helper_url, model, apiToken
}
