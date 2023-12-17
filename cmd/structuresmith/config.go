package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFile represents the structure of the configuration file.
type ConfigFile struct {
	TemplateGroups map[string][]FileStructure `yaml:"templateGroups"`
	Projects       []ProjectConfig            `yaml:"repositories"`
}

// ProjectConfig defines the configuration of a single repository.
type ProjectConfig struct {
	Name   string             `yaml:"name"`
	Files  []FileStructure    `yaml:"files"`
	Groups []TemplateGroupRef `yaml:"groups"`
}

// TemplateGroupRef links a template group with specific values.
type TemplateGroupRef struct {
	GroupName string                 `yaml:"groupName"`
	Values    map[string]interface{} `yaml:"values"`
}

// FileStructure describes a file to be created from a template or URL.
type FileStructure struct {
	Destination string `yaml:"destination"`
	Source      string `yaml:"source"`
	SourceURL   string `yaml:"sourceUrl"`
	Content     string `yaml:"content"`
	Values      map[string]interface{}
}

// Template represents a template consisting of multiple files.
type Template struct {
	Files []FileStructure
}

// TemplateGroup is a collection of templates.
type TemplateGroup []Template

// Project represents a repository with associated templates and groups.
type Project struct {
	Name   string
	Files  []FileStructure
	Groups []TemplateGroupRef
}

// readConfig reads and parses the YAML configuration file.
func readConfig(filename, templatesDir string) (ConfigFile, error) {
	var config ConfigFile
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Add templatesDir prefix to sources in template groups
	for _, group := range config.TemplateGroups {
		for i, file := range group {
			if file.Source != "" {
				group[i].Source = filepath.Join(templatesDir, file.Source)
			}
		}
	}

	// Add templatesDir prefix to sources directly under repositories
	for i, repo := range config.Projects {
		for j, file := range repo.Files {
			if file.Source != "" {
				config.Projects[i].Files[j].Source = filepath.Join(templatesDir, file.Source)
			}
		}
	}

	log.Println("Configuration read successfully.")
	return config, nil
}

// getProject checks if a repository exists in the given list of repositories.
func (c *ConfigFile) findProjectConfig(name string) (ProjectConfig, bool) {
	for _, project := range c.Projects {
		if project.Name == name {
			return project, true
		}
	}
	return ProjectConfig{}, false
}

// validateConfig performs various checks on the configuration.
func (c *ConfigFile) validateConfig() error {
	if err := c.validateDuplicateProjectNames(); err != nil {
		return err
	}
	if err := c.validateDuplicateTemplateGroups(); err != nil {
		return err
	}
	if err := c.validateFileStructures(); err != nil {
		return err
	}
	if err := c.validateProjectGroupReferences(); err != nil {
		return err
	}
	if err := c.validateURLSchemes(); err != nil {
		return err
	}
	return nil
}

// validateDuplicateProjectNames checks for duplicate repository names.
func (c *ConfigFile) validateDuplicateProjectNames() error {
	repoNames := make(map[string]bool)
	for _, repo := range c.Projects {
		if _, exists := repoNames[repo.Name]; exists {
			return fmt.Errorf("duplicate repository name: %s", repo.Name)
		}
		repoNames[repo.Name] = true
	}
	return nil
}

// validateDuplicateTemplateGroups checks for duplicate template groups.
func (c *ConfigFile) validateDuplicateTemplateGroups() error {
	groupNames := make(map[string]bool)
	for groupName := range c.TemplateGroups {
		if _, exists := groupNames[groupName]; exists {
			return fmt.Errorf("duplicate template group: %s", groupName)
		}
		groupNames[groupName] = true
	}
	return nil
}

// validateFileStructures checks for conflicts in file structures.
func (c *ConfigFile) validateFileStructures() error {
	for _, files := range c.TemplateGroups {
		for _, file := range files {
			if file.Source != "" && file.Content != "" {
				return fmt.Errorf("both SourceFile and Content set for file: %s", file.Destination)
			}
			if file.Source != "" {
				if _, err := os.Stat(file.Source); os.IsNotExist(err) {
					return fmt.Errorf("template file or directory not found: %s", file.Source)
				}
			}
		}
	}
	return nil
}

// validateProjectGroupReferences checks if repositories refer to valid groups.
func (c *ConfigFile) validateProjectGroupReferences() error {
	for _, repo := range c.Projects {
		for _, groupRef := range repo.Groups {
			if _, exists := c.TemplateGroups[groupRef.GroupName]; !exists {
				return fmt.Errorf("repository %s refers to non-existent group: %s", repo.Name, groupRef.GroupName)
			}
		}
	}
	return nil
}

// validateURLSchemes checks the validity of URLs in file structures.
func (c *ConfigFile) validateURLSchemes() error {
	for _, files := range c.TemplateGroups {
		for _, file := range files {
			if file.SourceURL != "" {
				if _, err := url.ParseRequestURI(file.SourceURL); err != nil {
					return fmt.Errorf("invalid SourceURL: %s", file.SourceURL)
				}
			}
		}
	}
	return nil
}

func (c *ConfigFile) FindProject(project string) (Project, error) {
	projectCfg, found := c.findProjectConfig(project)
	if !found {
		return Project{}, fmt.Errorf("project %s not found in configuration", project)
	}
	return Project(projectCfg), nil
}
