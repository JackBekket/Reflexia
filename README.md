# Reflexia

Reflexia is a code summarizer that can be used to generate summaries of code files in a project. It uses a combination of command-line arguments, environment variables, and configuration files to determine how to process the code. The project also includes a helper service that can be used to generate summaries of code snippets.

## Project Configuration

The project uses a configuration file to specify the settings for the summarizer. The configuration file can be specified using the `-l` flag. The configuration file can be in TOML format and contains the following fields:

- FileFilter: A list of file extensions to include in the summarization process.
- ProjectRootFilter: A list of file extensions to exclude from the summarization process.
- ModuleMatch: A regular expression to match module names.
- StopWords: A list of words to ignore during summarization.
- CodePrompt: A prompt to use when generating summaries of code snippets.
- SummaryPrompt: A prompt to use when generating summaries of text.
- PackagePrompt: A prompt to use when generating summaries of package-level code.
- ReadmePrompt: A prompt to use when generating summaries of README files.
- RootPath: The root directory of the project.

## Summarization Process

The summarization process begins by reading the configuration file and parsing the settings. The project then iterates through the files in the project directory, excluding any files specified in the `.gitignore` file. For each file, it checks if it matches any of the file filters specified in the configuration file. If a match is found, the file content is read, and a summary is generated using the appropriate prompt from the configuration file. The summaries are then stored in a map, where the key is the relative path to the file, and the value is the generated summary.

## Output

The project generates several output files:

- `FILES.md`: A markdown file containing the summaries of all files in the project.
- `README.md`: A markdown file containing the summaries of all README files in the project.
- `SUMMARY.md`: A markdown file containing the summaries of all files in the project.

The project also has the option to create a README file for each package in the project, as well as to overwrite the existing README file in the root project directory.

## Edge Cases

The project has several edge cases that can be handled using command-line arguments:

- `-c`: Do not check project root folder specific files such as go.mod or package.json.
- `-s`: Do not create SUMMARY.md and README.md, just print the file summaries.
- `-r`: Do not create README.md.
- `-p`: Do not create README.md for every package in the project.
- `-br`: Overwrite README.md for the root project directory instead of creating README_GENERATED.md.
- `-f`: Save individual file summary intermediate result to the FILES.md.
- `-bp`: Create README_GENERATED.md if README.md exists in the package directory instead of overwriting.

## Environment Variables

The project uses the following environment variables:

- `HELPER_URL`: Helper URL.
- `MODEL`: Model.
- `API_TOKEN`: API token.

## Run Instructions

1. Make sure you have the necessary dependencies installed.
2. Create a configuration file in TOML format and specify the settings for the summarizer.
3. Run the project with the desired command-line arguments and environment variables.

## Example

```bash
reflexia -l config.toml -c -s -r -p -br -f -bp
```



This command will run the Reflexia project using the configuration file `config.toml`, exclude project root folder specific files, not create SUMMARY.md and README.md, not create README.md for every package in the project, overwrite README.md for the root project directory instead of creating README_GENERATED.md, save individual file summary intermediate result to the FILES.md, and create README_GENERATED.md if README.md exists in the package directory instead of overwriting.


```bash
./reflexia -l go.toml -f -d  -g https://github.com/JackBekket/Hellper
```
This command will start to autodocumentize -g repository (public), it will create README_GENERATED.MD and create summary for all files
