# Reflexia

**Reflexia** is a tool for managing project workflows, including GitHub repository interactions, package processing, and integration with LLMs (e.g., OpenAI). It handles project configuration, code analysis, and GitHub PR creation based on configurable behavior.

## Package Structure
- `reflexia.go`: Contains the core implementation, including the `ReflexiaCall.Run` entry point.

## Configuration Options
All behavior is controlled via the `ReflexiaCall` struct, which includes:

### Core Options
- `LocalWorkdir` (string): Local working directory path.
- `RepositoryURL` (string): GitHub repository URL.
- `RepositoryBranch` (string): Branch to checkout.
- `GithubUsername` (string): GitHub username for authentication.
- `GithubToken` (string): GitHub API token.
- `WithConfigFile` (string): Path to config file.
- `ExactPackages` (string): Restrict to specific packages.
- `CreatePR` (bool): Create GitHub PR.
- `LightCheck` (bool): Skip some checks.
- `WithFileSummary` (bool): Generate file summaries.
- `UseEmbeddings` (bool): Enable embeddings support.
- `OverwriteReadme` (bool): Overwrite README.md.
- `OverwriteCache` (bool): Overwrite cache entries.

### Config Fields
- `Config.CachePath` (string): Path to cache directory.
- `Config.EmbeddingsAIURL` (string): Embeddings AI API endpoint.
- `Config.EmbeddingsAIToken` (string): Embeddings AI API token.
- `Config.EmbeddingsDBURL` (string): Database URL for embeddings.
- `Config.EmbeddingsSimSearchTestPrompt` (string): Test prompt for similarity search.

### AgentConfig Fields
- `AIURL` (string): AI API endpoint.
- `AIToken` (string): AI API token.
- `Model` (string): LLM model to use.

## Notes
- **Unimplemented/Undefined Behavior**: 
  - Invalid/missing configuration values (e.g., missing `GithubToken`) are not handled.
  - `OverwriteCache` behavior is not clearly documented.
  - Multiple project config variants are not fully explained.
- **Critical Points**: 
  - `ChooserFunc` for selecting project configs is a critical failure point.
  - `SimulateSearchTest` is deferred only if a test prompt is set.
- **Assumptions**: 
  - The project directory structure is valid.
  - GitHub PR creation does not handle API rate limiting or authentication errors.