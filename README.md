<h1 align="center">
<img src=".img/logo.png" width="125px"/>
 <br>
 structuresmith
 </br>
</h1>
<h4 align="center">Automates the generation of project files and templates for repositories (and others) based on customizable YAML configurations</h4>
<p align="center">
  <a href="https://github.com/cbrgm/structuresmith"><img src="https://img.shields.io/github/release/cbrgm/structuresmith.svg" alt="GitHub release"></a>
  <a href="https://goreportcard.com/report/github.com/cbrgm/structuresmith"><img src="https://goreportcard.com/badge/github.com/cbrgm/structuresmith" alt="Go Report Card"></a>
</p>

## Features 🌟

- **Template Generation** 📄: Generates files from templates for consistent repository setup.
- **YAML Configuration** ⚙️: Easy-to-define templates and file structures in YAML format.
- **Repository Customization** 📚: Unique configurations using a "mixin" approach for each repository.
- **Content Flexibility** ✍️: Supports external template files, direct content in YAML, and files from URLs.
- **Robust Validation** 🔍: Ensures configuration integrity with comprehensive validation checks before processing.
- **Selective Processing** 🎯: Processes a single or multiple repositories as needed.
- **Clean Output** 🧹: Clears existing output directories for up-to-date content generation.

## Installation

You can download the latest release of structuresmith with this one-liner on MacOS / Linux (amd64 + arm64):

```
wget -O structuresmith "https://github.com/cbrgm/structuresmith/releases/latest/download/structuresmith_$(uname -s)-$(uname -m)"
```

You may also download the latest pre-compiled binaries from the [GitHub releases page](https://github.com/cbrgm/structuresmith/releases/latest) or build `structuresmith` from source using `Go` (1.21+):

```bash
make build
```


## Usage

After installing, you can run Structuresmith with the following command-line arguments:

```bash
structuresmith --config path/to/config.yaml --output output/directory --templates path/to/templates
```

### Container

```bash
podman run --rm -it ghcr.io/cbrgm/structuresmith:latest
```

### GitHub Actions

Please check out the [action.yml](./action.yml) and the example [workflow](.github/workflows/example-workflow.yml).

## Command-Line Arguments

* `-c`, `--config`: Path to the YAML configuration file.
* `-o`, `--output`: Output path prefix for generated files. (default: `./out`)
* `-t`, `--templates`: Directory for template files. (default: `./templates`)
* `-r`, `--repo`: Specify a single repository to process.
* `-p`, `--max-parallel`: Maximum number of repositories to process in parallel. (default: `5`)

## Configuration Overview

Please take a look at the sample [configuration.yml](./configuration.yml) in this repository and check out the [template](./templates/) directoy.

```yaml
# Define template groups with a set of files
templateGroups:

  # Group for common git files like .gitignore and LICENSE
  commonGitFiles:
    - destination: ".gitignore"
      source: "gitignore.tmpl"  # Template for .gitignore file
    - destination: "LICENSE"
      source: "license.tmpl"    # Template for LICENSE file
    - destination: ".img/logo.png"  # Also supports binary or non-text copies
      source: "logo.png"

  # Group for Go-specific files like .golangci.yml
  projectSpecificFiles:
    - destination: ".golangci.yml"
      source: "golangci.tmpl"   # Template for .golangci.yml
      values:
        golangciVersion: "1.54.x"

    # Also supports downloading text and non-text / binary files from (accessible) URLs
    # You may also add values here used for templating after downloading the artifact
    - destination: "Dockerfile"
      sourceUrl: "https://raw.githubusercontent.com/cbrgm/promcheck/main/Dockerfile"

    # Group for templating a whole directory
    # You may also add values here as well for templating files while traversing the directory.
    - destination: "docs/"
      source: "docs_templaes/"  # Directory containing multiple template files
      values:
        foo: "bar"

# List of repositories / projects to apply the template groups
repositories:

  # First repository configuration
  - name: "example/repo1"  # This will create a subdirectory ./out/example/repo1/
    groups:
      - groupName: "commonGitFiles"   # Referencing commonGitFiles group
        values:
          Author: "Author Name"       # Values to substitute in templates
          Year: "2023"
      - groupName: "projectSpecificFiles"  # Referencing projectSpecificFiles group

  # Second repository configuration
  - name: "example/repo2"
    groups:
      - groupName: "commonGitFiles"
        values:
          Author: "Chris Bargmann"
          Year: "2023"

  # Third repository with individual files but no groups
  - name: "some-project"  # This will create a subdirectory ./out/some-project/
    files:
      - destination: "README.md"
        content: |
          # Welcome to some-project
          This repository contains various examples.
      - destination: "config.json"
        content: |
          {
            "version": "1.0",
            "description": "Configuration file for some-project"
          }
```

### Example 1: Basic Repository with Common and Project Specific Files

```yaml
templateGroups:
  # [Existing commonGitFiles and projectSpecificFiles definitions]

repositories:
  - name: "basic-project"
    groups:
      - groupName: "commonGitFiles"
        values:
          Author: "Jane Doe"
          Year: "2023"
      - groupName: "projectSpecificFiles"
```

### Example 2: Advanced Repository with Custom Content and Directory Templating

```yaml
templateGroups:
  # [Existing templateGroups definitions]

repositories:
  - name: "advanced-custom-project"
    files:
      - destination: "special.md"
        content: |
          # Special Markdown File
          This is a specially crafted markdown file.
    groups:
      - groupName: "projectSpecificFiles"
      - groupName: "commonGitFiles"
        values:
          Author: "Alice Johnson"
          Year: "2023"
```

### Example 3: Repository with External Resources and Nested Templating Values

```yaml
templateGroups:
  # [Existing templateGroups definitions]

repositories:
  - name: "external-resource-project"
    files:
      - destination: "external_documentation.md"
        sourceUrl: "https://example.com/documentation.md"
        values:
          Author: "Bob Smith"
          Year: "2023"
          Nested:
            Level1: "Value1"
            Level2:
              Key: "Value2"
```

### Example 4: Repository with Direct Directory Structure
```yaml
templateGroups:
  # [Existing templateGroups definitions]

repositories:
  - name: "direct-directory-structure-project"
    files:
      - destination: "docs/"
        source: "documentation_templates/"  # Source directory containing documentation templates
      - destination: "scripts/"
        source: "script_templates/"          # Source directory for script templates
        values:
            UseBash: true
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

## Contributing & License

We welcome and value your contributions to this project! 👍 If you're interested in making improvements or adding features, please refer to our [Contributing Guide](https://github.com/cbrgm/semver-bump-action/blob/main/CONTRIBUTING.md). This guide provides comprehensive instructions on how to submit changes, set up your development environment, and more.

Please note that this project is developed in my spare time and is available for free 🕒💻. As an open-source initiative, it is governed by the [Apache 2.0 License](https://github.com/cbrgm/semver-bump-action/blob/main/LICENSE). This license outlines your rights and obligations when using, modifying, and distributing this software.

Your involvement, whether it's through code contributions, suggestions, or feedback, is crucial for the ongoing improvement and success of this project. Together, we can ensure it remains a useful and well-maintained resource for everyone 🌍.
