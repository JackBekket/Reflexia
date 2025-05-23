**packagerunner**

**Description**  
The `packagerunner` package provides functionality to analyze and generate summaries for project packages, handling code analysis, documentation generation, and integration with LLM-based tools. It processes packages, generates file and package-level summaries, manages fallback behaviors, and interacts with project configuration and external services.

**External Data/Config Options**  
- **ProjectConfig**:  
  - `Prompts`: Map of model names to prompt configurations:  
    - `CodePrompt`: Prompt for file code analysis.  
    - `CodePromptFallback`: Fallback prompt for code analysis.  
    - `PackagePrompt`: Prompt for package analysis.  
    - `PackagePromptFallback`: Fallback prompt for package analysis.  
  - `RootPath`: Project root directory path.  

**Behavior Control**  
- `PkgFiles`: Map of package to file paths (controls which files are processed).  
- `ProjectConfig`: Defines project-specific behavior (e.g., prompts, root path).  
- `SummarizeService`: Handles LLM-based content generation.  
- `EmbeddingsService`: Manages document storage via embeddings.  
- `ExactPackages`: Restricts processing to specific packages.  
- `OverwriteReadme`: Controls behavior of `README.md` generation.  
- `WithFileSummary`: Enables `FILES.md` generation.  
- `Model`: Specifies the target LLM model.  
- `PrintTo`: Redirects logs to a writer.  

**Edge Cases/Notes**  
- Packages with no files are skipped.  
- Empty file content results in a warning.  
- Failed LLM requests propagate errors.  
- Invalid project configurations (e.g., missing default prompts) return errors.  
- Fallback mechanisms are used for empty summaries.  
- Special handling exists for `README.md` generation.  

**Dependencies**  
- LLM integration via `SummarizeService.LLMRequest`.  
- File system operations for reading/writing files.  
- `EmbeddingsService.Store` for document storage.  

No explicit TODOs in code. Error handling is present for file operations and invalid configurations.