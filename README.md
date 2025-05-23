# Reflexia

A tool for analyzing and managing software projects, providing both an API and CLI interface for handling project configurations, code analysis, and integration with GitHub.

---

## 📦 Project Structure

- **`cmd/api`**: API server entry point (`api.go`). Handles HTTP API endpoints.
- **`cmd/cli`**: CLI tool entry point (`main.go`). Executes command-line operations.
- **`internal/`**: Internal logic (e.g., `github/pull_request.go`, `project/choosers.go`).
- **`pkg/`**: Core functionality (e.g., `config`, `project`, `store`, `summarize`).
- **`project_config/`**: Language-specific configuration files (TOML: `cpp.toml`, `go.toml`, etc.).

---

## 🛠 Behavior Control

### 🔧 Configuration Options

**Loaded via `pkg/config`**: Values are defined in project-specific TOML files (e.g., `project_config/go.toml`).  
**Effects**:
- Determines supported languages and analysis rules.
- Controls project handling and GitHub integration behavior.
- **No explicit validation** of configuration values in the provided code.

### ⚙ Build Tags

- **`api`**: Enables API server functionality (excludes CLI code).
- **`!api`**: Disables API server (only CLI is active).

---

## 🧩 External Configuration Options

### 📦 Environment Variables

| Variable | Description | Notes |
|--------|-------------|-------|
| `GH_TOKEN` | GitHub authentication token (overrides `-t` flags) | CLI only |
| `LISTEN_ADDR` | API server listen address (required) | API only |
| `CORS_ALLOW_ORIGINS` | Comma-separated list of allowed CORS origins (required) | API only |

### 🛠 Command-line Flags (CLI Only)

| Flag | Description |
|------|-------------|
| `-a` | Cache folder path (default: `.reflexia_cache`) |
| `-eu` | Embeddings AI URL |
| `-ea` | Embeddings AI API Key |
| `-ed` | Embeddings DB connect URL |
| `-et` | Embeddings similarity test prompt |
| `-g` | GitHub repository URL |
| `-b` | GitHub repository branch |
| `-u` | GitHub username for SSH auth |
| `-t` | GitHub token for SSH auth (overrides `GH_TOKEN`) |
| `-l` | Config filename in `project_config` to use |
| `-p` | Exact package names (comma delimited) |
| `-w` | Create PR (implies `ReflexiaOpts.CreatePR = true`) |
| `-c` | Skip project root checks (implies `ReflexiaOpts.LightCheck = true`) |
| `-f` | Save file summaries to `FILES.md` (implies `ReflexiaOpts.WithFileSummary = true`) |
| `-r` | Overwrite `README.md` (implies `ReflexiaOpts.OverwriteReadme = true`) |
| `-d` | Overwrite cache (implies `ReflexiaOpts.OverwriteCache = true`) |
| `-e` | Use embeddings (implies `ReflexiaOpts.UseEmbeddings = true`) |

---

## ⚠ Limitations & Notes

### General
- **No error handling** for failed configuration loading or API startup.
- **Potential panics** if `config.NewConfig()` returns invalid data.
- **Unaddressed edge cases**: Behavior when configuration files are missing or malformed.

### CLI Specific
- `printEmptyWarning` only outputs warnings for **non-empty** response lists.
- `agentConfig` is **separate** from the main config and is loaded via `agentConfig.NewConfig()`.
- **Edge Cases**: 
  - If `RepositoryURL` is empty and command-line arguments exist, the first argument is treated as a local working directory.
  - Invalid URLs or missing `go.toml` files in projects will cause failures.
- **Test Issues**: 
  - `TestProcessWorkingDirectory` has a duplicate test case.
  - No handling of invalid project directories in core functions (only tested via edge cases).
- **Dead Code**: 
  - `TestProcessWorkingDirectoryIgnore` is not implemented in the core logic.

### API Specific
- API endpoints `/project_configs` and `/reflect` are not fully documented in OpenAPI.
- No error handling for invalid OpenAPI routes.
- Workdir is determined at startup and used by the APIService.
- Environment variables `LISTEN_ADDR` and `CORS_ALLOW_ORIGINS` are critical; missing values cause panic.

---

## 📌 Usage

### CLI
Run the CLI to analyze projects, interact with GitHub, or manage configurations:
```bash
go build . -o reflexia
reflexia [flags] [args]
```

### API
Start the API server with:
```bash
go build -tags=api . -o reflexia-api
reflexia-api
```
Ensure environment variables `LISTEN_ADDR` and `CORS_ALLOW_ORIGINS` are set.

---

## 📝 Project Goals

- Provide a unified interface for project analysis and GitHub integration.
- Support multiple languages via project-specific configurations.
- Enable both CLI and API usage for flexibility.

---

## 🧩 Project Configurations

Project-specific configurations are stored in `project_config/` as TOML files (e.g., `go.toml`, `cpp.toml`). These files define:
- Supported languages.
- Analysis rules.
- Project handling behavior.
- GitHub integration policies.

---

## 📦 Project Scope

This project is intended for developers who want to:
- Analyze and manage software projects.
- Integrate with GitHub for code analysis and PR automation.
- Use both CLI and API interfaces for project management tasks.

---

## 📦 Project Status

This project is in early development. It provides basic functionality for project analysis and GitHub integration, but lacks full error handling, comprehensive documentation, and complete test coverage. Contributions are welcome.
