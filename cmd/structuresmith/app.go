package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

// Structuresmith holds configuration paths for the application.
type Structuresmith struct {
	ConfigFile   string
	OutputDir    string
	TemplatesDir string
}

// Options represents the command line arguments passed to Structuresmith.
type Options struct {
	ConfigFile   string
	OutputDir    string
	TemplatesDir string
}

// newStructuresmith initializes a new instance of Structuresmith with provided options.
func newStructuresmith(opts Options) *Structuresmith {
	return &Structuresmith{
		ConfigFile:   opts.ConfigFile,
		OutputDir:    opts.OutputDir,
		TemplatesDir: opts.TemplatesDir,
	}
}

// loadAndValidateConfig loads and validates the configuration file.
func (app *Structuresmith) loadAndValidateConfig() (ConfigFile, error) {
	log.Println("Reading configuration...")
	config, err := readConfig(app.ConfigFile, app.TemplatesDir)
	if err != nil {
		return ConfigFile{}, fmt.Errorf("error reading config: %v", err)
	}

	if err := config.validateConfig(); err != nil {
		return ConfigFile{}, err
	}
	return config, nil
}

// diff generates a diff of the project file structures.
func (app *Structuresmith) diff(project string, cfg ConfigFile) error {
	projectConfig, err := cfg.FindProject(project)
	if err != nil {
		return err
	}

	allFiles, err := app.processProject(projectConfig, cfg.TemplateGroups)
	if err != nil {
		return err
	}

	lock, err := LoadOrCreateLockFile(app.OutputDir)
	if err != nil {
		return err
	}

	diffedFiles := lock.Diff(allFiles)
	fmt.Printf("\n%s\n", diffedFiles)
	return nil
}

func (app *Structuresmith) render(project string, cfg ConfigFile) error {
	p, err := cfg.FindProject(project)
	if err != nil {
		return err
	}

	allFiles, err := app.processProject(p, cfg.TemplateGroups)
	if err != nil {
		return err
	}

	lock, err := LoadOrCreateLockFile(app.OutputDir)
	if err != nil {
		return err
	}

	diffedFiles := lock.Diff(allFiles)
	fmt.Printf("\n%s\n", diffedFiles)

	for _, file := range allFiles {
		if err = app.renderFileStructure(file); err != nil {
			return err
		}
	}

	err = app.deleteOrphanedFileStructures(diffedFiles)
	if err != nil {
		return err
	}

	err = WriteLockFile(allFiles, app.OutputDir)
	if err != nil {
		return err
	}

	return nil
}

// renderFileStructure creates a file based on the FileStructure details.
func (app *Structuresmith) renderFileStructure(file FileStructure) error {
	outputPath := filepath.Join(app.OutputDir, filepath.Dir(file.Destination))
	fullPath := filepath.Join(outputPath, filepath.Base(file.Destination))

	log.Printf("Processing %s", fullPath)
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Handle different file sources
	switch {
	case file.Content != "":
		return handleFileCreation(fullPath, file.Content, file.Values)
	case file.SourceURL != "":
		content, err := downloadFileContent(file.SourceURL)
		if err != nil {
			return fmt.Errorf("downloading file from URL: %w", err)
		}
		return handleFileCreation(fullPath, content, file.Values)
	case file.Source != "":
		content, err := os.ReadFile(file.Source)
		if err != nil {
			return fmt.Errorf("reading source file: %w", err)
		}
		return handleFileCreation(fullPath, string(content), file.Values)
	default:
		return fmt.Errorf("file structure lacks source information")
	}
}

// handleFileCreation creates or copies a file based on provided content or source.
func handleFileCreation(fullPath, content string, values map[string]interface{}) error {
	// Attempt to create a templated file
	err := createTemplatedFile(fullPath, content, values)
	if err != nil {
		// If templating fails, copy the content directly
		return copyContentToFile(content, fullPath)
	}
	return nil
}

// deleteOrphanedFileStructures removes any files that are no longer needed.
func (app *Structuresmith) deleteOrphanedFileStructures(diffResult DiffResult) error {
	for _, file := range diffResult.DeletedFiles {
		fullPath := filepath.Join(app.OutputDir, file.Destination)
		log.Printf("Deleting %s", fullPath)

		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing file %s: %w", fullPath, err)
		}

		if err := app.removeEmptyDirs(filepath.Dir(fullPath)); err != nil {
			return err
		}
	}
	return nil
}

// removeEmptyDirs recursively removes empty directories.
func (app *Structuresmith) removeEmptyDirs(dir string) error {
	for dir != app.OutputDir {
		files, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("error reading directory %s: %w", dir, err)
		}

		if len(files) == 0 {
			if err := os.Remove(dir); err != nil {
				return fmt.Errorf("error removing directory %s: %w", dir, err)
			}
			dir = filepath.Dir(dir)
		} else {
			break
		}
	}
	return nil
}

func (app *Structuresmith) processProject(p Project, globalGroups map[string][]FileStructure) ([]FileStructure, error) {
	var allFiles []FileStructure

	// Process individual files
	for _, file := range p.Files {
		files, err := app.processFileStructure(file)
		if err != nil {
			return nil, fmt.Errorf("error processing file structure: %w", err)
		}
		allFiles = append(allFiles, files...)
	}

	// Process groups of files
	for _, groupRef := range p.Groups {
		group, exists := globalGroups[groupRef.GroupName]
		if !exists {
			return nil, fmt.Errorf("template group %s not found in configuration", groupRef.GroupName)
		}

		for _, file := range group {
			mergedValues := mergeValues(file.Values, groupRef.Values)
			file.Values = mergedValues
			files, err := app.processFileStructure(file)
			if err != nil {
				return nil, fmt.Errorf("error processing file structure: %w", err)
			}
			allFiles = append(allFiles, files...)
		}
	}

	return allFiles, nil
}

// mergeValues merges group-level values with file-level.
// For slices and non-map values, the value from src will overwrite the one in dst.
func mergeValues(dst, src map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy values from dst into the new map
	for key, val := range dst {
		merged[key] = val
	}

	// Merge values from src into the new map
	for key, srcVal := range src {
		if dstVal, exists := merged[key]; exists {
			// If both values are maps, merge them recursively
			if srcMap, srcOk := srcVal.(map[string]interface{}); srcOk {
				if dstMap, dstOk := dstVal.(map[string]interface{}); dstOk {
					merged[key] = mergeValues(dstMap, srcMap)
					continue
				}
			}
		}
		// For all other cases, or if the key doesn't exist in merged, set/overwrite the merged value
		merged[key] = srcVal
	}
	return merged
}

// processFileStructure determines whether the source is a file or directory and processes accordingly.
func (app *Structuresmith) processFileStructure(file FileStructure) ([]FileStructure, error) {
	if file.Source != "" {
		fileInfo, err := os.Stat(file.Source)
		if err != nil {
			return nil, fmt.Errorf("error stating source: %w", err)
		}

		if fileInfo.IsDir() {
			return app.processDirectory(file)
		}
		return []FileStructure{file}, nil
	}

	if file.SourceURL != "" || file.Content != "" {
		return []FileStructure{file}, nil
	}

	return nil, fmt.Errorf("neither source, sourceUrl nor content defined in file structure: %v", file)
}

// processDirectory processes each file within a directory.
func (app *Structuresmith) processDirectory(directory FileStructure) ([]FileStructure, error) {
	var allFiles []FileStructure
	err := filepath.Walk(directory.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(directory.Source, path)
			if err != nil {
				return fmt.Errorf("error getting relative path: %w", err)
			}
			allFiles = append(allFiles, FileStructure{
				Source:      path,
				Destination: filepath.Join(directory.Destination, relPath),
				Values:      directory.Values,
			})
		}
		return nil
	})
	return allFiles, err
}

// downloadFileContent fetches content from a URL and returns it as a string.
func downloadFileContent(fileURL string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("error closing response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("bad response status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}

// createTemplatedFile creates a file from a template content and values.
func createTemplatedFile(path, content string, values map[string]interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Printf("error closing file: %v\n", cerr)
		}
	}()

	tmpl, err := template.New(filepath.Base(path)).Parse(content)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := tmpl.Execute(f, values); err != nil {
		return err // Return the error to indicate templating failure
	}

	return nil
}

// copyFile copies a file from source to destination.
// nolint: unused
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading source file for copy: %w", err)
	}

	if err := os.WriteFile(dst, input, 0o644); err != nil {
		return fmt.Errorf("writing copied file: %w", err)
	}
	return nil
}

// copyContentToFile writes string content directly to a file.
func copyContentToFile(content, filePath string) error {
	return os.WriteFile(filePath, []byte(content), 0o644)
}
