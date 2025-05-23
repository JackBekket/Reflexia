# Package summarize

## Overview
The `summarize` package provides functionality for managing text summarization workflows using LLM (Large Language Model) integration, caching, and text processing. It enables generating summaries by combining LLM outputs with cache handling and post-processing logic.

## File Structure
- `summarize.go`: Contains the core implementation for summarization, caching, and LLM interaction.

## External Configuration Options
The package behavior is influenced by the following external data and configuration options:

| Option              | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `Agent`             | `agent.Agent` interface for interacting with the agent system.                |
| `LlmOptions`        | `[]llms.CallOption` options passed to the LLM invocation.                    |
| `OverwriteCache`    | `bool`: Bypasses cache and overwrites existing entries.                      |
| `IgnoreCache`       | `bool`: Disables cache usage entirely.                                      |
| `CachePath`         | `string`: Directory path for cache files (invalid paths may cause failures).  |
| `StopWords`         | `[]string`: Suffixes to trim from LLM responses (only exact matches).        |
| `Model`             | `string`: LLM model identifier (affects cache key validity).                 |

## Notes
- **Cache Behavior**: Cache operations may fail if `CachePath` is inaccessible or invalid. Cache entries are invalidated by changes to `Model` or the formatted prompt.
- **Security**: `saveCache` creates directories with `os.ModePerm` (world-writable), which may pose security risks.
- **Edge Cases**: 
  - `StopWords` trimming is non-atomic (only removes exact suffix matches).
  - Error handling in `loadCache` ignores `fs.ErrNotExist` but returns other errors.
- **No Explicit TODOs**: The code contains no TODO comments, but potential improvements include handling cache path errors and refining `StopWords` logic.