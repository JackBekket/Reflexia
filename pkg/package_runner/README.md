# Package: packagerunner

This package provides a service for running packages and generating summaries for them. It utilizes various external services and data sources to achieve this.

### Imports:

- context
- fmt
- io/fs
- log
- os
- path/filepath
- slices
- strings
- github.com/JackBekket/reflexia/internal
- github.com/JackBekket/reflexia/pkg
- github.com/JackBekket/reflexia/pkg/project
- github.com/JackBekket/reflexia/pkg/summarize
- github.com/tmc/langchaingo/schema

### External Data and Input Sources:

- Project configuration: The package relies on a project configuration, which is likely stored in a file or database. This configuration contains information about the project, such as the root path, code prompt, and package prompt.
- Embeddings service: The package uses an embeddings service to store and retrieve embeddings for code and summaries. This service is likely an external API or a local database.
- Summarize service: The package uses a summarize service to generate summaries for code and packages. This service is likely an external API or a local service.

### Major Code Parts:

1. PackageRunnerService: This struct holds the necessary data and services for running packages and generating summaries. It includes:
    - PkgFiles: A map of package names to a list of file paths.
    - ProjectConfig: A pointer to the project configuration.
    - SummarizeService: A pointer to the summarize service.
    - EmbeddingsService: A pointer to the embeddings service.
    - ExactPackages: A string containing a comma-separated list of package names to include.
    - OverwriteReadme: A boolean indicating whether to overwrite existing README files.
    - WithFileSummary: A boolean indicating whether to generate a FILES.md file for each package.

2. RunPackages: This function iterates through the PkgFiles map and runs each package. For each package, it:
    - Generates a summary for each file in the package.
    - Generates a summary for the entire package.
    - Writes the summaries to the appropriate files (README.md and FILES.md).
    - Stores the embeddings for the summaries in the embeddings service.

3. getDirFileStructure: This function walks through a directory and returns a string containing the file structure of the directory.

4. fileMapToMd: This function takes a map of file paths to summaries and returns a string containing the summaries in Markdown format.

5. fileMapToString: This function takes a map of file paths to summaries and returns a string containing the summaries.

6. writeFile: This function takes a file path and content and writes the content to the file.

7. getReadmePath: This function determines the path to the README file for a given directory.



