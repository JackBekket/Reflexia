# internal/api

This package provides HTTP endpoints for handling reflection and project configuration operations.

## Features

- `ReflectPost` (POST endpoint): Handles reflection operations with various configuration options.
- `ProjectConfigsGet` (GET endpoint): Retrieves project configuration mappings.

## Configuration Options

### `ReflectInput` (all optional except `RepositoryURL`):
- `AIURL`, `AIToken`, `Model`: AI system configuration.
- `RepositoryURL`: Required repository address.
- `RepositoryBranch`: Optional branch to analyze.
- `GithubUsername`, `GithubToken`: GitHub authentication credentials.
- `WithConfigFile`: Config file path to use.
- `ExactPackages`: Packages to restrict analysis to.
- `CreatePR`, `LightCheck`, `WithFileSummary`: Analysis behavior flags.
- `OverwriteReadme`, `OverwriteCache`: Cache/README handling flags.
- `UseEmbeddings`: Whether to use embeddings for analysis.

## Notes

- `RepositoryURL` is required; empty value returns an error.
- `PullRequestURL` is optional and only set if non-nil.
- Project config conversion only handles `FileFilter` and `ProjectRootFilter` fields.
- `ReflexiaCall` uses `project.FirstChooser` as the default package selector.