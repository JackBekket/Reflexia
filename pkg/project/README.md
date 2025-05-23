# project

The `project` package handles project configuration loading and management, providing tools to select between multiple configurations based on project files and environmental constraints.

## File Structure

- **choosers.go**: Contains configuration selection logic (`FirstChooser`, `CLIChooser`).
- **project.go**: Manages `ProjectConfig` loading, validation, and file processing.

## What It Does

This package provides:
- Loading and parsing of project configurations from TOML files.
- Mechanisms to choose between multiple project configurations (e.g., first available, CLI interaction).
- Handling of project-specific logic like file filtering, directory grouping, and package management.

## Configuration

### External Data/Configuration Options
- **TOML Files**: Define `ProjectConfig` behavior (loaded into struct fields).
- **ProjectConfig Fields**:
  - `FileFilter []string`: File suffix filters (e.g., `.go`).
  - `ProjectRootFilter []string`: Root directory filters (e.g., `go.mod`).
  - `ModuleMatch string`: Must be `directory` or `go_package` (controls file grouping).
  - `StopWords []string`: Words to ignore during processing.
  - `Prompts map[string]ProjectConfigPrompts`: Custom prompts for code/package handling.
  - `RootPath string`: Project root directory path.

### Supported Languages
- Hardcoded support for `go`, `python`, `typescript` (in error messages).

## Notes

- `ProjectConfig` type is **not defined in this package** (assumed to be defined elsewhere).
- `CLIChooser` has **unhandled edge cases** (e.g., invalid inputs, infinite loops).
- `FirstChooser` returns the first sorted key from `projectConfigVariants` (behavior not fully explained in code).
- `BuildPackageFiles()` **fails** if `ModuleMatch` is not `directory` or `go_package`.
- `ProjectConfig` fields like `Prompts` have **fallback behavior** (not explicitly documented).