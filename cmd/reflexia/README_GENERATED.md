## Package: reflexia

This package provides a tool for summarizing code and generating documentation for software projects. It supports multiple programming languages and can be used to create summaries of individual files or entire packages.

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
    - `GH_TOKEN`: GitHub token for SSH authentication
    - `CACHE_PATH`: Cache folder path
    - `EMBEDDINGS_AI_URL`: Embeddings AI URL
    - `EMBEDDINGS_AI_KEY`: Embeddings AI API Key
    - `EMBEDDINGS_DB_URL`: Embeddings pgxpool DB connect URL
    - `EMBEDDINGS_SIM_SEARCH_TEST_PROMPT`: Embeddings similarity search validation test prompt
- Command-line arguments:
    - `-g`: Valid link for GitHub repository
    - `-u`: GitHub username for SSH authentication
    - `-t`: GitHub token for SSH authentication
    - `-l`: Config filename in project_config to use
    - `-a`: Cache folder path
    - `-eu`: Embeddings AI URL
    - `-ea`: Embeddings AI API Key
    - `-ed`: Embeddings pgxpool DB connect URL
    - `-et`: Embeddings similarity search validation test prompt
    - `-p`: Exact package names, comma-delimited
    - `-c`: Do not check project root folder specific files
    - `-f`: Save individual file summary intermediate result to the FILES.md
    - `-r`: Overwrite README.md instead of README_GENERATED.md creation/overwrite
    - `-d`: Overwrite generated summary caches
    - `-e`: Use Embeddings

### Major Code Parts:

1. Config initialization:
    - Loads environment variables using `godotenv.Load()`.
    - Initializes a `Config` struct with default values for various parameters.
    - Parses command-line arguments using `flag.Parse()`.

2. Process working directory:
    - Determines the working directory based on the provided GitHub link, username, and token.
    - If no GitHub link is provided, it uses the first command-line argument as the working directory.
    - If the working directory is a GitHub repository, it clones the repository using `git.PlainClone()`.

3. Choose project config:
    - Determines the appropriate project configuration based on the detected programming language.
    - If multiple configurations are found, it prompts the user to choose one.

4. Run packages:
    - Creates a `PackageRunnerService` instance with the necessary parameters.
    - Calls the `RunPackages()` method to process the packages and generate summaries.

5. Print empty warnings:
    - Prints warnings if any fallback attempts or empty LLM responses were encountered during the process.

6. Similarity search test:
    - If embeddings are enabled, it performs a similarity search using the provided test prompt and prints the results.

7. Print results:
    - Prints the generated summaries and other relevant information.

### Summary:

This package provides a comprehensive solution for summarizing code and generating documentation for software projects. It supports multiple programming languages, allows for customization through command-line arguments and environment variables, and provides various options for handling different scenarios. The package is well-structured and easy to use, making it a valuable tool for developers and project managers alike.

cmd/reflexia/reflexia.go
reflexia.go
reflexia_test.go