<h1 align="center">
<img src=".img/logo.png" width="125px"/>
 <br>
 structuresmith
 </br>
</h1>
<h4 align="center">Structuresmith is a powerful tool designed to automate the generation of project files, streamlining repository setup and more using customizable templates. It's ideal for developers looking to maintain consistency and efficiency in their project configurations.</h4>
<p align="center">
  <a href="https://github.com/cbrgm/structuresmith"><img src="https://img.shields.io/github/release/cbrgm/structuresmith.svg" alt="GitHub release"></a>
  <a href="https://goreportcard.com/report/github.com/cbrgm/structuresmith"><img src="https://goreportcard.com/badge/github.com/cbrgm/structuresmith" alt="Go Report Card"></a>
  <a href="https://github.com/cbrgm/structuresmith/actions/workflows/go-lint-test.yml"><img src="https://github.com/cbrgm/structuresmith/actions/workflows/go-lint-test.yml/badge.svg" alt="go-lint-test"></a>
  <a href="https://github.com/cbrgm/structuresmith/actions/workflows/go-binaries.yml"><img src="https://github.com/cbrgm/structuresmith/actions/workflows/go-binaries.yml/badge.svg" alt="go-binaries"></a>
  <a href="https://github.com/cbrgm/structuresmith/actions/workflows/container.yml"><img src="https://github.com/cbrgm/structuresmith/actions/workflows/container.yml/badge.svg" alt="container"></a>
</p>

- [Features 🌟](#features-)
- [What can this tool do for you?](#what-can-this-tool-do-for-you)
- [Installation](#installation)
- [CLI Usage](#cli-usage)
   * [Validate](#validate)
   * [Diff](#diff)
   * [Render](#render)
- [Container](#container)
- [GitHub Actions](#github-actions)
- [Configuration Overview](#configuration-overview)
- [Examples](#examples)
   * [Example 1: Simple Inline Content for a Single Project](#example-1-simple-inline-content-for-a-single-project)
   * [Example 2: Single File Template with Value](#example-2-single-file-template-with-value)
   * [Example 3: Templated File with Nested Values](#example-3-templated-file-with-nested-values)
   * [Example 4: Using Template Group with Overwritten Values](#example-4-using-template-group-with-overwritten-values)
   * [Example 5: Multiple Template Groups](#example-5-multiple-template-groups)
   * [Example 6: Mix of Direct Files and Template Groups](#example-6-mix-of-direct-files-and-template-groups)
   * [Example 7: Templating a Whole Directory](#example-7-templating-a-whole-directory)
   * [Example 8: Downloading Content from URLs](#example-8-downloading-content-from-urls)
- [Lockfile `.anvil.lock`](#lockfile-anvillock)
- [Templating Explained](#templating-explained)
   * [How It Works](#how-it-works)
   * [Go Templating Syntax](#go-templating-syntax)
- [Contributing & License](#contributing-license)

---

## Features 🌟

- **Template Generation**: Automates the creation of project files, ensuring consistency and standardization across different projects
- **YAML Configuration**: Easily define templates and file structures in YAML format.
- **Robust Validation**: Ensures configuration integrity with comprehensive checks and file diffing before processing.
- **Content Flexibility**:  Supports a variety of content sources, including external templates, direct YAML content, and files from URLs, making it adaptable to different project requirements.

## What can this tool do for you?

1. **Automated Project Setup**: Streamline new project initialization with predefined templates, significantly reducing setup time.
2. **Standardization Across Projects**: Ensure consistent file structures and setups across multiple projects or repositories, crucial for team coherence and project maintenance.
3. **Bootstrapping**: Accelerate the creation of initial project structures for quick prototyping and development iterations.
4. **Custom Template Management**: Efficiently manage and apply custom templates across various projects, especially beneficial in large teams or organizations.
5. **Configuration Auditing**: Validate YAML configurations before deployment to ensure compliance with required standards and specifications, enhancing reliability and reducing configuration errors.

## Installation

You can download the latest release of structuresmith with this one-liner on MacOS / Linux (amd64 + arm64):

```
wget -O structuresmith "https://github.com/cbrgm/structuresmith/releases/latest/download/structuresmith_$(uname -s)-$(uname -m)"
```

You may also download the latest pre-compiled binaries from the [GitHub releases page](https://github.com/cbrgm/structuresmith/releases/latest) or build `structuresmith` from source using `Go` (1.21+):

```bash
make build
```

## CLI Usage

After installing, you can run Structuresmith with the following command-line arguments:

```bash
structuresmith -h
```

- `-h, --help`: Shows context-sensitive help. This flag can be used with any command to get more information about its usage and options.
- `--config="anvil.yml"`: Specifies the path to the YAML configuration file. This flag allows you to define a custom configuration file for the tool to use.
- `--output="out"`: Sets the output path prefix for the generated files. This flag lets you specify where the generated files should be stored.
- `--templates="templates"`: Indicates the directory where template files are stored. With this flag, you can define a custom location for your template files.

### Validate

Validates the YAML configuration (`anvil.yml`) to ensure its integrity and checks for any potential issues.

```bash
structuresmith validate --config path/to/config.yaml
```

### Diff

Conducts a dry-run to display the file paths that would be generated, helping to preview changes without actual file creation.

```bash
structuresmith diff --config path/to/config.yaml --output output/directory --templates path/to/templates project-to-render
```

Example output:

```bash
delete:     .gitignore
overwrite:  .golangci.yml
overwrite:  Dockerfile
overwrite:  LICENSE
new:        foobar.txt
overwrite:  sub/bar.txt
overwrite:  sub/foo.txt
delete:     sub/nested/foo.txt
```

### Render

Processes and writes the templated files to the disk, applying the configurations to generate the specified project structure.

```bash
structuresmith render --config path/to/config.yaml --output output/directory --templates path/to/templates project-to-render
```

## Container

```bash
podman run --rm -it ghcr.io/cbrgm/structuresmith:latest
```

## GitHub Actions

Please check out the [action.yml](./action.yml) and the example [workflow](.github/workflows/example-workflow.yml).

## Configuration Overview

These examples showcase the versatility and capabilities of `structuresmith`, ranging from simple inline content to more complex configurations using template groups and nested values. Please take a look at the sample [anvil.yml](./anvil.yml).

```yaml
# Structuresmith YAML Configuration
# This configuration showcases the use of template groups, custom values for templating,
# and the definition of project-specific files. Modify paths and values as needed.

# Define projects to apply template groups
projects:

  # Example Go project configuration
  - name: "example/go-project"
    groups:
      - groupName: "commonFiles"
      - groupName: "goProjectFiles"
        values:
          packageName: "main"   # Custom value used in Go template

  # Example for a general project using inline content
  - name: "example/general-project"
    groups:
      - groupName: "commonFiles"
    files:
      - destination: "config.json"
        content: |
          {
            "setting": "value",
            "enabled": true
          }

# Define template groups with sets of files
templateGroups:

  # Common files for all projects
  commonFiles:
    - destination: ".gitignore"
      source: "templates/gitignore.tmpl"  # Template for .gitignore
    - destination: "README.md"
      source: "templates/readme.tmpl"     # Template for README.md

# Group for specific project types, e.g., Go projects
  goProjectFiles:
    - destination: "main.go"
      source: "templates/main.go.tmpl"    # Main file for Go project
    - destination: "Makefile"
      source: "templates/Makefile.tmpl"   # Makefile for build commands

# Uncomment to demonstrate downloading files from URLs
#   - destination: "Dockerfile"
#     sourceUrl: "https://example.com/Dockerfile"
# Uncomment to demonstrate copying whole directories
#   - destination: "docs/"
#     source: "docs_templates/"

```

## Examples

### Example 1: Simple Inline Content for a Single Project

**Description**: A basic configuration creating a `README.md` file with inline content.

**YAML Configuration**:
```yaml
projects:
  - name: "simple-project"
    files:
      - destination: "README.md"
        content: "Welcome to Simple Project"
```

**Output:**

* `out/README.md` containing "Welcome to Simple Project".

### Example 2: Single File Template with Value

**Description**: Creating a `config.txt` file from a template, substituting a value.
**YAML Configuration**:
```yaml
templateGroups:
  configFile:
    - destination: "config.txt"
      source: "templates/config.tmpl"
projects:
  - name: "config-project"
    groups:
      - groupName: "configFile"
        values:
          setting: "Enabled"
```

**Output:**

* `out/config.txt` with content from `config.tmpl`, where a template value like `{{ .setting }}` is replaced by "Enabled".

### Example 3: Templated File with Nested Values

**Description**: Creating a file with nested template values.
**YAML Configuration**:
```yaml
templateGroups:
  detailFile:
    - destination: "details.txt"
      source: "templates/details.tmpl"
projects:
  - name: "detail-project"
    groups:
      - groupName: "detailFile"
        values:
          user:
            name: "Alice"
            role: "Developer"

```

**Output:**

* `out/details.txt` with content from `details.tmpl`, where template values like `{{ .user.name }}` are replaced by "Alice" and `{{ .user.role }}` by "Developer".

### Example 4: Using Template Group with Overwritten Values

**Description**: Utilizing a template group with values defined in the group and overwritten in the project definition.
**YAML Configuration**:
```yaml
templateGroups:
  baseFiles:
    - destination: "base.txt"
      source: "templates/base.tmpl"
      values:
        defaultText: "Default"
projects:
  - name: "base-project"
    groups:
      - groupName: "baseFiles"
        values:
          defaultText: "Customized Text"
```

**Output:**

* `out/base.txt` with content from `base.tmpl`, where the default text is overwritten by "Customized Text".

### Example 5: Multiple Template Groups

**Description**: Combining multiple template groups in a single project.
**YAML Configuration**:
```yaml
templateGroups:
  commonFiles:
    - destination: "README.md"
      source: "templates/readme.tmpl"
  additionalFiles:
    - destination: "extra.txt"
      source: "templates/extra.tmpl"
projects:
  - name: "multi-group-project"
    groups:
      - groupName: "commonFiles"
      - groupName: "additionalFiles"
```

**Output:**

* `out/README.md` from `readme.tmpl`.
* `out/extra.txt` from `extra.tmpl`.

### Example 6: Mix of Direct Files and Template Groups

**Description**: A project configuration using a mix of direct file definitions and template groups.
**YAML Configuration**:
```yaml
templateGroups:
  documentationFiles:
    - destination: "docs/intro.md"
      source: "templates/docs/intro.tmpl"
projects:
  - name: "mixed-project"
    files:
      - destination: "overview.txt"
        content: "Project Overview"
    groups:
      - groupName: "documentationFiles"
```

**Output:**

* `out/overview.txt` with "Project Overview".
* `out/docs/intro.md` generated from intro.tmpl.

### Example 7: Templating a Whole Directory

**Description**: Applying templates to an entire directory.
**YAML Configuration**:
```yaml
projects:
  completeDirectory:
    - destination: "config/"
      source: "templates/config_directory/"
      values:
        appName: "MyApp"

repositories:
  - name: "directory-project"
    groups:
      - groupName: "completeDirectory"
```

**Output:**

* Directory `out/config/` with files from `templates/config_directory/`, where template values like `{{ .appName }}` are replaced with "MyApp".

### Example 8: Downloading Content from URLs

**Description**: Fetching a file from a URL and placing it into the project directory.
**YAML Configuration**:
```yaml
projects:
  - name: "download-project"
    files:
      - destination: "Dockerfile"
        sourceUrl: "https://raw.githubusercontent.com/exampleuser/project/main/Dockerfile"
```

**Output:**

* `out/Dockerfile` containing the content fetched from the provided URL.

## Lockfile `.anvil.lock`

Structuresmith's `anvil.lock` file is vital for managing project files. It keeps a record of used files and templates, tracking updates since the last use of the tool. An important feature of Structuresmith is its ability to automatically remove files from the project's output directory that are no longer present in the original project configuration. This ensures the output remains synchronized with the current project setup.

Including `anvil.lock` in the project's versioning is beneficial. It provides a clear history of file changes, especially important in team settings to maintain consistency and prevent conflicts in the project's files.

## Templating Explained

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
