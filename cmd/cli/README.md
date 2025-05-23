# cli

This package provides the command-line interface (CLI) for the Reflexia tool, enabling project analysis, code summarization, and interaction with GitHub repositories. It handles command-line flags, configuration management, and integration with LLMs (Large Language Models) for code-related tasks.

## File Structure
- `cli.go`: Contains the core logic for handling command-line operations, including the `Run` function, `printEmptyWarning`, and `readReflexiaCall`.
- `cli_test.go`: Contains test functions for handling working directories, project configurations, and LLM interactions.

## Configuration
### Environment Variables
- `GH_TOKEN`: Used for GitHub authentication (overrides `-t` flags).

### Command-line Flags
- `-a`: Cache folder path (default: `.reflexia_cache`)
- `-eu`: Embeddings AI URL
- `-ea`: Embeddings AI API Key
- `-ed`: Embeddings DB connect URL
- `-et`: Embeddings similarity test prompt
- `-g`: GitHub repository URL
- `-b`: GitHub repository branch
- `-u`: GitHub username for SSH auth
- `-t`: GitHub token for SSH auth (overrides `GH_TOKEN`)
- `-l`: Config filename in project_config to use
- `-p`: Exact package names (comma delimited)
- `-w`: Create PR (implies `ReflexiaOpts.CreatePR = true`)
- `-c`: Skip project root checks (implies `ReflexiaOpts.LightCheck = true`)
- `-f`: Save file summaries to FILES.md (implies `ReflexiaOpts.WithFileSummary = true`)
- `-r`: Overwrite README.md (implies `ReflexiaOpts.OverwriteReadme = true`)
- `-d`: Overwrite cache (implies `ReflexiaOpts.OverwriteCache = true`)
- `-e`: Use embeddings (implies `ReflexiaOpts.UseEmbeddings = true`)

## Notes
- The `Run` function exits with a **fatal error** on critical failures.
- `printEmptyWarning` only outputs warnings for **non-empty** response lists.
- The `agentConfig` is **separate** from the main config and is loaded via `agentConfig.NewConfig()`.
- **Edge Cases**: 
  - If `RepositoryURL` is empty and command-line arguments exist, the first argument is treated as a local working directory.
  - Invalid URLs or missing `go.toml` files in projects will cause failures.
- **Test Issues**: 
  - `TestProcessWorkingDirectory` has a duplicate test case.
  - No handling of invalid project directories in core functions (only tested via edge cases).
- **Dead Code**: 
  - `TestProcessWorkingDirectoryIgnore` is not implemented in the core logic.
