## Package: reflexia

This package provides a tool for summarizing code and generating README files for Go, Python, and TypeScript projects. It leverages embeddings and similarity search to enhance the summarization process.

### Imports:

- `context`
- `errors`
- `flag`
- `fmt`
- `log`
- `net/url`
- `os`
- `path/filepath`
- `strconv`
- `strings`
- `github.com/go-git/go-git/v5`
- `github.com/go-git/go-git/v5/plumbing/transport/http`
- `github.com/joho/godotenv`
- `github.com/tmc/langchaingo/llms`
- `github.com/JackBekket/reflexia/pkg`
- `github.com/JackBekket/reflexia/pkg/package_runner`
- `github.com/JackBekket/reflexia/pkg/project`
- `github.com/JackBekket/reflexia/pkg/summarize`

### External Data and Input Sources:

- Environment variables:
    - `GH_TOKEN`: GitHub token for SSH authentication.
    - `CACHE_PATH`: Cache folder path (defaults to ".reflexia_cache").
    - `EMBEDDINGS_AI_URL`: Embeddings AI URL.
    - `EMBEDDINGS_AI_KEY`: Embeddings AI API key.
    - `EMBEDDINGS_DB_URL`: Embeddings pgxpool DB connect URL.
    - `EMBEDDINGS_SIM_SEARCH_TEST_PROMPT`: Embeddings similarity search validation test prompt.
- Command-line flags:
    - `-g`: Valid link for GitHub repository.
    - `-u`: GitHub username for SSH authentication.
    - `-t`: GitHub token for SSH authentication.
    - `-l`: Config filename in project_config to use.
    - `-a`: Cache folder path.
    - `-eu`: Embeddings AI URL.
    - `-ea`: Embeddings AI API key.
    - `-ed`: Embeddings pgxpool DB connect URL.
    - `-et`: Embeddings similarity search validation test prompt.
    - `-p`: Exact package names, comma-delimited.
    - `-c`: Do not check project root folder specific files.
    - `-f`: Save individual file summary intermediate result to the FILES.md.
    - `-r`: Overwrite README.md instead of README_GENERATED.md creation/overwrite.
    - `-d`: Overwrite generated summary caches.
    - `-e`: Use Embeddings.

### Code Summary:

1. **Initialization and Configuration:**
    - The code initializes a configuration struct and parses command-line flags and environment variables to set up the necessary parameters for the summarization process.

2. **Working Directory Setup:**
    - The code determines the working directory based on the provided GitHub link or command-line arguments. If a GitHub link is provided, it clones the repository to the specified working directory.

3. **Project Configuration Selection:**
    - The code identifies the appropriate project configuration based on the detected project language (Go, Python, or TypeScript) and user input.

4. **Embeddings Service Initialization (Optional):**
    - If the `UseEmbeddings` flag is set, the code initializes an embeddings service using the provided Embeddings AI URL, API key, and database URL.

5. **Package File Generation and Summarization:**
    - The code generates a list of package files based on the selected project configuration and runs the summarization process for each package.

6. **Similarity Search Validation (Optional):**
    - If an Embeddings similarity search test prompt is provided, the code performs a similarity search using the embeddings service and prints the results.

7. **Output and Reporting:**
    - The code prints any empty or fallback responses for files and packages, as well as any warnings related to the summarization process.

8. **Cache Management:**
    - The code overwrites the generated summary caches if the `OverwriteCache` flag is set.

9. **README Generation (Optional):**
    - If the `WithFileSummary` flag is set, the code saves individual file summaries to a FILES.md file.

10. **README Overwrite (Optional):**
    - If the `OverwriteReadme` flag is set, the code overwrites the README.md file instead of creating a new README_GENERATED.md file.



reflexia/
├── .env
├── __debug_bin2596789548
├── reflexia.go
└── reflexia_test.go
cmd/reflexia/
└── reflexia.go