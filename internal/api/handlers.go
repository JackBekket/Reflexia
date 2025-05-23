package api

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/JackBekket/reflexia/pkg/config"
	"github.com/JackBekket/reflexia/pkg/project"
	"github.com/JackBekket/reflexia/pkg/reflexia"
	agentConfig "github.com/Swarmind/libagent/pkg/config"
	"github.com/swaggest/usecase/status"
)

type APIService struct {
	Workdir string
	Config  config.Config
}

type ReflectInput struct {
	AIURL   string `json:"ai_url"`
	AIToken string `json:"ai_token"`
	Model   string `json:"model"`

	RepositoryURL    string `json:"repository_url"`
	RepositoryBranch string `json:"repository_branch,omitempty"`
	GithubUsername   string `json:"github_username,omitempty"`
	GithubToken      string `json:"github_token,omitempty"`

	WithConfigFile string `json:"with_config_file,omitempty"`
	ExactPackages  string `json:"exact_packages,omitempty"`

	CreatePR        bool `json:"create_pr,omitempty"`
	LightCheck      bool `json:"light_check,omitempty"`
	WithFileSummary bool `json:"with_file_summary,omitempty"`
	OverwriteReadme bool `json:"overwrite_readme,omitempty"`

	OverwriteCache bool `json:"overwrite_cache,omitempty"`
	UseEmbeddings  bool `json:"use_embeddings,omitempty"`
}

type ReflectOutput struct {
	PullRequestURL string `json:"pull_request_url"`
}

func (s APIService) ReflectPost(ctx context.Context,
	input ReflectInput,
	output *ReflectOutput,
) error {
	if input.RepositoryURL == "" {
		return status.Wrap(errors.New("empty repository_url"), status.InvalidArgument)
	}

	reflexiaCall := reflexia.ReflexiaCall{
		Config:      s.Config,
		ChooserFunc: project.FirstChooser,
		PrintTo:     io.Discard,

		AgentConfig: agentConfig.Config{
			AIURL:   input.AIURL,
			AIToken: input.AIToken,
			Model:   input.Model,
		},

		RepositoryURL:    input.RepositoryURL,
		RepositoryBranch: input.RepositoryBranch,
		GithubUsername:   input.GithubUsername,
		GithubToken:      input.GithubToken,
		WithConfigFile:   input.WithConfigFile,
		ExactPackages:    input.ExactPackages,
		CreatePR:         input.CreatePR,
		LightCheck:       input.LightCheck,
		WithFileSummary:  input.WithFileSummary,
		UseEmbeddings:    input.UseEmbeddings,
		OverwriteReadme:  input.OverwriteReadme,
		OverwriteCache:   input.OverwriteCache,
	}

	artifacts, err := reflexiaCall.Run(ctx)
	if err != nil {
		return status.Wrap(fmt.Errorf("reflexia call: %w", err), status.Internal)
	}

	if artifacts.PullRequestURL != nil {
		*output = ReflectOutput{
			PullRequestURL: *artifacts.PullRequestURL,
		}
	}

	return nil
}

type ProjectConfigsInput struct {
}

type ProjectConfig struct {
	FileFilter        []string `json:"file_filter"`
	ProjectRootFilter []string `json:"project_root_filter"`
}

func (s APIService) ProjectConfigsGet(ctx context.Context,
	input ProjectConfigsInput,
	output *map[string]ProjectConfig,
) error {
	pcs, err := project.ListProjectConfigs(s.Workdir)
	if err != nil {
		return status.Wrap(fmt.Errorf("listing project configs: %w", err), status.Internal)
	}

	resp := map[string]ProjectConfig{}
	for k, v := range pcs {
		resp[k] = ProjectConfig{
			FileFilter:        v.FileFilter,
			ProjectRootFilter: v.ProjectRootFilter,
		}
	}
	*output = resp

	return nil
}
