<h1 align="center">
<img src=".img/logo.png" width="175px"/>
 <br>
 structuresmith
 </br>
</h1>
<h4 align="center">Automates the generation of project files and templates for repositories (and others) based on customizable YAML configurations</h4>
<p align="center">
  <a href="https://github.com/cbrgm/structuresmith"><img src="https://img.shields.io/github/release/cbrgm/structuresmith.svg" alt="GitHub release"></a>
  <a href="https://goreportcard.com/report/github.com/cbrgm/structuresmith"><img src="https://goreportcard.com/badge/github.com/cbrgm/structuresmith" alt="Go Report Card"></a>
  <a href="https://github.com/cbrgm/structuresmith/actions/workflows/go-build.yml"><img src="https://github.com/cbrgm/structuresmith/actions/workflows/go-build.yml/badge.svg" alt="test-and-build"></a>
</p>

## Features 🌟

- **Template Generation** 📄: Generates files from templates for consistent repository setup.
- **YAML Configuration** ⚙️: Easy-to-define templates and file structures in YAML format.
- **Repository Customization** 📚: Unique configurations using a "mixin" approach for each repository.
- **Content Flexibility** ✍️: Supports external template files, direct content in YAML, and files from URLs.
- **Robust Validation** 🔍: Ensures configuration integrity with comprehensive validation checks before processing.
- **Selective Processing** 🎯: Processes a single or multiple repositories as needed.
- **Clean Output** 🧹: Clears existing output directories for up-to-date content generation.

## Usage

Before using Structuresmith, download the latest pre-compiled binaries from the [GitHub releases page](https://github.com/cbrgm/structuresmith/releases).

After downloading, you can run Structuresmith with the following command-line arguments:

```bash
structuresmith --config path/to/config.yaml --output output/directory --templates path/to/templates
```

## Command-Line Arguments

* `-c`, `--config`: Path to the YAML configuration file.
* `-o`, `--output`: Output path prefix for generated files. (default: `./out`)
* `-t`, `--templates`: Directory for template files. (default: `./templates`)
* `-r`, `--repo`: Specify a single repository to process.
* `-p`, `--max-parallel`: Maximum number of repositories to process in parallel. (default: `5`)

## Sample configurations

Please take a look at the sample [configuration.yml](./configuration.yml) in this repository and check out the [template](./templates/) directoy.

```yaml
# Define template groups with a set of files
templateGroups:

  # Group for common git files like .gitignore and LICENSE
  commonGitFiles:
    - filename: ".gitignore"
      sourceFile: "gitignore.tmpl"  # Template for .gitignore file
    - filename: "LICENSE"
      sourceFile: "license.tmpl"    # Template for LICENSE file
    - filename: ".img/logo.png"     # Also supports binary or non-text copies
      sourceFile: "logo.png"

    # Also supports downloading text and non-text / binary files from (accessible) URLs
    # You may also add values here used for templating after downloading the artifact
    - filename: "Dockerfile"
      sourceUrl: "https://raw.githubusercontent.com/cbrgm/promcheck/main/Dockerfile"
    - filename: "archive.tar.gz"
      sourceUrl: "https://github.com/cbrgm/promcheck/releases/download/v1.1.8/promcheck_darwin_amd64.tar.gz"

  # Group for Go-specific files like .golangci.yml
  goSpecificFiles:
    - filename: ".golangci.yml"
      sourceFile: "golangci.tmpl"   # Template for .golangci.yml
      values:
        golangciVersion: "1.54.x"

# List of repositories / projects to apply the template groups
repositories:

  # First repository configuration
  - name: "example/repo1" # This will create a subdirectory ./out/example/repo1/
    groups:
      - groupName: "commonGitFiles"   # Referencing commonGitFiles group
        values:
          Author: "Author Name"       # Values to substitute in templates
          Year: "2023"
      - groupName: "goSpecificFiles"  # Referencing goSpecificFiles group

  # Second repository configuration
  - name: "example/repo2"
    groups:
      - groupName: "commonGitFiles"
        values:
          Author: "Chris Bargmann"
          Year: "2023"

  # Third repository with individual files but no groups
  - name: "some-project" # This will create a subdirectory ./out/some-project/
    files:
      - filename: "README.md"
        content: |
          # Welcome to example/repo3
          This repository contains various examples.
        # Inline content is used here instead of a template file
      - filename: "config.json"
        content: |
          {
            "version": "1.0",
            "description": "Configuration file for example/repo3"
          }
```

### Kickstart configuration

To get you started: A minimal configuration with all file content in `yaml` may look like this:

```yaml
repositories:
  - name: "some-project"
    files:
      - filename: "README.md"
        content: |
          # Welcome to some-project, {{ .user }}!
          This repository contains various examples.
        values:
          user: "Bob"
      - filename: "config.json"
        content: |
          {
            "version": "1.0",
            "description": "Configuration file for some-project"
          }
```
Render all files to `./out/some-project/`:

```bash
structuresmith --config ./examples/configuration-simple.yml
```

Please help us adding more [examples](./examples).

## Output Directory Structure

When Structuresmith generates files, it creates an output directory structure that mirrors the structure defined in your YAML configuration. For each repository, a separate folder is created within the specified output directory.

## Templating Explained 📝

Structuresmith leverages Go's powerful templating system, allowing you to define dynamic content in your templates. This system provides a flexible way to insert values into your files, making your templates adaptable to different contexts.

### How It Works

- **Placeholders**: In your templates, you can use placeholders in the form `{{ .VariableName }}` to insert values dynamically. These placeholders will be replaced with actual values when the template is processed.
- **YAML Configuration**: Define the values for these placeholders in the YAML configuration under the `values` key for each template group or individual file.
- **Example**: If your template contains `{{ .Author }}`, you can specify the author's name in the YAML configuration, and it will be replaced in the generated file.

### Go Templating Syntax

- Structuresmith uses Go's `text/template` package syntax. This includes conditional statements, range loops, and more, providing a rich set of features for creating complex templates.
- For detailed information on Go's templating syntax, visit the official documentation: [Go's text/template package](https://golang.org/pkg/text/template/)


### Local Development

You can build `structuresmith` from source using `Go` (1.21+):

```bash
make build
```

## Contributing & License

Feel free to submit changes! See the [Contributing Guide](https://github.com/cbrgm/contributing/blob/master/CONTRIBUTING.md). This project is open-source
and is developed under the terms of the [Apache 2.0 License](https://github.com/cbrgm/structuresmith/blob/master/LICENSE).
