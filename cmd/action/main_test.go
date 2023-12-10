package main

import (
	"os"
	"reflect"
	"testing"
)

func TestValidateDuplicateRepoNames(t *testing.T) {
	tests := []struct {
		name    string
		repos   []RepositoryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "No Repositories",
			repos:   []RepositoryConfig{},
			wantErr: false,
		},
		{
			name:    "Single Repository",
			repos:   []RepositoryConfig{{Name: "repo1"}},
			wantErr: false,
		},
		{
			name:    "Two Different Repositories",
			repos:   []RepositoryConfig{{Name: "repo1"}, {Name: "repo2"}},
			wantErr: false,
		},
		{
			name:    "Duplicate Repositories",
			repos:   []RepositoryConfig{{Name: "repo1"}, {Name: "repo1"}},
			wantErr: true,
			errMsg:  "duplicate repository name: repo1",
		},
		{
			name:    "Multiple Repositories with One Duplicate",
			repos:   []RepositoryConfig{{Name: "repo1"}, {Name: "repo2"}, {Name: "repo1"}},
			wantErr: true,
			errMsg:  "duplicate repository name: repo1",
		},
		{
			name:    "Multiple Unique Repositories",
			repos:   []RepositoryConfig{{Name: "repo1"}, {Name: "repo2"}, {Name: "repo3"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDuplicateRepoNames(tt.repos)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDuplicateRepoNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("validateDuplicateRepoNames() got = %v, want %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateDuplicateTemplateGroups(t *testing.T) {
	testCases := []struct {
		name    string
		groups  map[string][]FileStructure
		wantErr bool
		errMsg  string
	}{
		{
			name:    "No groups",
			groups:  map[string][]FileStructure{},
			wantErr: false,
		},
		{
			name: "Single group",
			groups: map[string][]FileStructure{
				"group1": {},
			},
			wantErr: false,
		},
		{
			name: "Multiple unique groups",
			groups: map[string][]FileStructure{
				"group1": {},
				"group2": {},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDuplicateTemplateGroups(tc.groups)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s: expected error but got none", tc.name)
				} else if err.Error() != tc.errMsg {
					t.Errorf("%s: expected error message '%s', but got '%s'", tc.name, tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %s", tc.name, err)
				}
			}
		})
	}
}

func TestValidateFileStructures(t *testing.T) {
	// Create a temporary file for testing file existence
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testCases := []struct {
		name    string
		groups  map[string][]FileStructure
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty groups",
			groups:  map[string][]FileStructure{},
			wantErr: false,
		},
		{
			name: "Valid file structures",
			groups: map[string][]FileStructure{
				"group1": {
					{Filename: "file1", SourceFile: tempFile.Name()},
				},
			},
			wantErr: false,
		},
		{
			name: "Both SourceFile and Content set",
			groups: map[string][]FileStructure{
				"group1": {
					{Filename: "file1", SourceFile: "file1.tmpl", Content: "content"},
				},
			},
			wantErr: true,
			errMsg:  "both SourceFile and Content set for file: file1",
		},
		{
			name: "SourceFile not found",
			groups: map[string][]FileStructure{
				"group1": {
					{Filename: "file2", SourceFile: "nonexistent.tmpl"},
				},
			},
			wantErr: true,
			errMsg:  "template file not found: nonexistent.tmpl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFileStructures(tc.groups)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s: expected error but got none", tc.name)
				} else if err.Error() != tc.errMsg {
					t.Errorf("%s: expected error message '%s', but got '%s'", tc.name, tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %s", tc.name, err)
				}
			}
		})
	}
}

func TestValidateRepoGroupReferences(t *testing.T) {
	testCases := []struct {
		name    string
		repos   []RepositoryConfig
		groups  map[string][]FileStructure
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid group references",
			repos:   []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "group1"}}}},
			groups:  map[string][]FileStructure{"group1": {}},
			wantErr: false,
		},
		{
			name:    "Non-existent group reference",
			repos:   []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "nonexistent"}}}},
			groups:  map[string][]FileStructure{"group1": {}},
			wantErr: true,
			errMsg:  "repository repo1 refers to non-existent group: nonexistent",
		},
		{
			name:    "Multiple repositories with valid references",
			repos:   []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "group1"}}}, {Name: "repo2", Groups: []TemplateGroupRef{{GroupName: "group2"}}}},
			groups:  map[string][]FileStructure{"group1": {}, "group2": {}},
			wantErr: false,
		},
		{
			name:    "Multiple repositories with a mix of valid and invalid references",
			repos:   []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "group1"}}}, {Name: "repo2", Groups: []TemplateGroupRef{{GroupName: "nonexistent"}}}},
			groups:  map[string][]FileStructure{"group1": {}},
			wantErr: true,
			errMsg:  "repository repo2 refers to non-existent group: nonexistent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRepoGroupReferences(tc.repos, tc.groups)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s: expected error but got none", tc.name)
				} else if err.Error() != tc.errMsg {
					t.Errorf("%s: expected error message '%s', but got '%s'", tc.name, tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %s", tc.name, err)
				}
			}
		})
	}
}

func TestValidateURLSchemes(t *testing.T) {
	testCases := []struct {
		name    string
		groups  map[string][]FileStructure
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid URL",
			groups:  map[string][]FileStructure{"group1": {{SourceURL: "http://example.com/file"}}},
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			groups:  map[string][]FileStructure{"group1": {{SourceURL: "invalid-url"}}},
			wantErr: true,
			errMsg:  "invalid SourceURL: invalid-url",
		},
		{
			name:    "Empty URL",
			groups:  map[string][]FileStructure{"group1": {{SourceURL: ""}}},
			wantErr: false,
		},
		{
			name:    "Multiple URLs with mixed validity",
			groups:  map[string][]FileStructure{"group1": {{SourceURL: "http://valid.com"}, {SourceURL: "bad-url"}}},
			wantErr: true,
			errMsg:  "invalid SourceURL: bad-url",
		},
		{
			name:    "No URL field",
			groups:  map[string][]FileStructure{"group1": {{Filename: "file1"}}},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateURLSchemes(tc.groups)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s: expected error but got none", tc.name)
				} else if err.Error() != tc.errMsg {
					t.Errorf("%s: expected error message '%s', but got '%s'", tc.name, tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %s", tc.name, err)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid configuration",
			config:  Config{}, // A valid minimal configuration
			wantErr: false,
		},
		{
			name:    "Duplicate repository names",
			config:  Config{Repositories: []RepositoryConfig{{Name: "repo1"}, {Name: "repo1"}}},
			wantErr: true,
			errMsg:  "duplicate repository name: repo1",
		},
		{
			name:    "Invalid file structure - both SourceFile and Content set",
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Filename: "file1", SourceFile: "file1.tmpl", Content: "content"}}}},
			wantErr: true,
			errMsg:  "both SourceFile and Content set for file: file1",
		},
		{
			name:    "Invalid file structure - non-existent SourceFile",
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Filename: "file1", SourceFile: "nonexistent.tmpl"}}}},
			wantErr: true,
			errMsg:  "template file not found: nonexistent.tmpl",
		},
		{
			name:    "Invalid group reference in repository",
			config:  Config{Repositories: []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "nonexistent"}}}}, TemplateGroups: map[string][]FileStructure{"group1": {}}},
			wantErr: true,
			errMsg:  "repository repo1 refers to non-existent group: nonexistent",
		},
		{
			name:    "Invalid URL scheme",
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Filename: "file1", SourceURL: "invalid-url"}}}},
			wantErr: true,
			errMsg:  "invalid SourceURL: invalid-url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.config)
			if tc.wantErr {
				if err == nil {
					t.Errorf("%s: expected error but got none", tc.name)
				} else if err.Error() != tc.errMsg {
					t.Errorf("%s: expected error message '%s', but got '%s'", tc.name, tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %s", tc.name, err)
				}
			}
		})
	}
}

func TestMergeValues(t *testing.T) {
	testCases := []struct {
		name        string
		groupValues map[string]interface{}
		fileValues  map[string]interface{}
		expected    map[string]interface{}
	}{
		{
			name:        "Both maps empty",
			groupValues: map[string]interface{}{},
			fileValues:  map[string]interface{}{},
			expected:    map[string]interface{}{},
		},
		{
			name:        "Group map empty",
			groupValues: map[string]interface{}{},
			fileValues:  map[string]interface{}{"key1": "value1"},
			expected:    map[string]interface{}{"key1": "value1"},
		},
		{
			name:        "File map empty",
			groupValues: map[string]interface{}{"key1": "value1"},
			fileValues:  map[string]interface{}{},
			expected:    map[string]interface{}{"key1": "value1"},
		},
		{
			name:        "No overlap",
			groupValues: map[string]interface{}{"key1": "groupValue1"},
			fileValues:  map[string]interface{}{"key2": "fileValue2"},
			expected:    map[string]interface{}{"key1": "groupValue1", "key2": "fileValue2"},
		},
		{
			name:        "Overlap - group values take precedence",
			groupValues: map[string]interface{}{"key1": "groupValue1"},
			fileValues:  map[string]interface{}{"key1": "fileValue1", "key2": "fileValue2"},
			expected:    map[string]interface{}{"key1": "groupValue1", "key2": "fileValue2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mergeValues(tc.groupValues, tc.fileValues)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}
