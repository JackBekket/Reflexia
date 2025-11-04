## Package: `util`

This package provides a utility function, `WalkDirIgnored`, for walking directories while respecting `.gitignore` patterns. It's designed for file system traversal with exclusion capabilities. The core functionality revolves around recursively visiting files and directories within a specified working directory (`workdir`), optionally filtering entries based on rules defined in an optional `.gitignore` file (`gitignorePath`).

**Configuration:**

*   `workdir`: (string) Root directory to traverse. Required.
*   `gitignorePath`: (string, optional) Path to the `.gitignore` file. If provided, files matching its patterns are excluded.
*   `WalkDirIgnoredFunction f`: Callback function executed for each non-ignored entry.

**Workflow:**

1.  The `WalkDirIgnored` function starts at the specified `workdir`.
2.  If a `gitignorePath` is provided, it attempts to load and compile the `.gitignore` file using `github.com/crackcomm/go-gitignore`. Errors during loading are logged but don't halt execution.
3.  The function recursively walks through subdirectories, skipping the `.git` directory for performance reasons.
4.  For each entry (file or directory), it checks if it matches any of the patterns in the loaded `.gitignore` file (if provided).
5.  If an entry is not ignored, the callback function `f` is invoked with the path to that entry.

**Dependencies:**

*   `github.com/crackcomm/go-gitignore`: For parsing and matching against `.gitignore` patterns.
*   `github.com/rs/zerolog/log`: Used for structured logging of errors (e.g., when loading `.gitignore`).

**File Structure:**

```
internal/util/
├── util.go
```