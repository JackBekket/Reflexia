# util

The `util` package provides utilities for directory traversal while respecting `.gitignore` rules.

## File Structure

- `util.go`: Contains the `WalkDirIgnored` function and related utilities.

## External Configuration

- `gitignorePath`: Path to a `.gitignore` file (optional). If provided, the package will load and respect the ignore rules.
- `.gitignore` files: Loaded via `ignore.CompileIgnoreFile`. The package does not handle invalid `.gitignore` files.

## TODOs & Comments

- Limited error handling for `gitignoreFile`: Only non-`os.ErrNotExist` errors are handled.
- No explicit handling for invalid `.gitignore` files.

## Edge Cases

- If `gitignorePath` is empty, no `.gitignore` processing occurs.
- `.git` directories are always skipped, overriding any `.gitignore` rules.
- The `WalkDirIgnoredFunction` may return `filepath.SkipDir` to abort directory traversal.
- `MatchesPath` checks relative paths, requiring `filepath.Rel(workdir, path)`.