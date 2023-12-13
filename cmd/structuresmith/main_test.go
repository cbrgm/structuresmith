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
					{Destination: "file1", Source: tempFile.Name()},
				},
			},
			wantErr: false,
		},
		{
			name: "Both SourceFile and Content set",
			groups: map[string][]FileStructure{
				"group1": {
					{Destination: "file1", Source: "file1.tmpl", Content: "content"},
				},
			},
			wantErr: true,
			errMsg:  "both SourceFile and Content set for file: file1",
		},
		{
			name: "SourceFile not found",
			groups: map[string][]FileStructure{
				"group1": {
					{Destination: "file2", Source: "nonexistent.tmpl"},
				},
			},
			wantErr: true,
			errMsg:  "template file or directory not found: nonexistent.tmpl",
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
			groups:  map[string][]FileStructure{"group1": {{Destination: "file1"}}},
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
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Destination: "file1", Source: "file1.tmpl", Content: "content"}}}},
			wantErr: true,
			errMsg:  "both SourceFile and Content set for file: file1",
		},
		{
			name:    "Invalid file structure - non-existent SourceFile",
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Destination: "file1", Source: "nonexistent.tmpl"}}}},
			wantErr: true,
			errMsg:  "template file or directory not found: nonexistent.tmpl",
		},
		{
			name:    "Invalid group reference in repository",
			config:  Config{Repositories: []RepositoryConfig{{Name: "repo1", Groups: []TemplateGroupRef{{GroupName: "nonexistent"}}}}, TemplateGroups: map[string][]FileStructure{"group1": {}}},
			wantErr: true,
			errMsg:  "repository repo1 refers to non-existent group: nonexistent",
		},
		{
			name:    "Invalid URL scheme",
			config:  Config{TemplateGroups: map[string][]FileStructure{"group1": {{Destination: "file1", SourceURL: "invalid-url"}}}},
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

func TestMergeValuesRecursively(t *testing.T) {
	tests := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Simple merge",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Overwrite value",
			dst: map[string]interface{}{
				"key": "original",
			},
			src: map[string]interface{}{
				"key": "new",
			},
			expected: map[string]interface{}{
				"key": "new",
			},
		},
		{
			name: "Merge nested maps",
			dst: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
				},
			},
			src: map[string]interface{}{
				"nested": map[string]interface{}{
					"key2": "value2",
				},
			},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "Merge with empty src map",
			dst: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			src: map[string]interface{}{},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Merge with empty dst map",
			dst:  map[string]interface{}{},
			src: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Nested maps with overwriting",
			dst: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "original",
					"key2": "keep",
				},
			},
			src: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "new",
					"key3": "add",
				},
			},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "new",
					"key2": "keep",
					"key3": "add",
				},
			},
		},
		{
			name: "Nested maps with different levels",
			dst: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "value",
				},
			},
			src: map[string]interface{}{
				"level1": "new value",
			},
			expected: map[string]interface{}{
				"level1": "new value",
			},
		},
		{
			name: "Mixed types in nested maps",
			dst: map[string]interface{}{
				"key1": map[string]interface{}{
					"nestedKey": 10,
				},
				"key2": "string",
			},
			src: map[string]interface{}{
				"key1": map[string]interface{}{
					"nestedKey": "overwritten",
				},
				"key2": 20,
			},
			expected: map[string]interface{}{
				"key1": map[string]interface{}{
					"nestedKey": "overwritten",
				},
				"key2": 20,
			},
		},
		{
			name: "Complex nested structure with maps and slices",
			dst: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2a": map[string]interface{}{
						"key1": "old1",
						"key2": []interface{}{"item1", "item2"},
						"key3": "old3",
					},
					"level2b": map[string]interface{}{
						"nestedInLevel2b": "oldValue",
					},
				},
				"level1b": []interface{}{
					map[string]interface{}{
						"arrayKey1": "arrayValue1",
					},
					"arrayItem2",
				},
			},
			src: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2a": map[string]interface{}{
						"key1": "new1",
						// key2 is a slice, should overwrite
						"key2": []interface{}{"item3"},
						"key3": "new3",
					},
					// level2b is a map, should merge
					"level2b": map[string]interface{}{
						"nestedInLevel2b": "newValue",
					},
				},
				// level1b is a slice, should overwrite
				"level1b": []interface{}{
					map[string]interface{}{
						"arrayKey1": "arrayValue1Modified",
					},
					"arrayItem3",
				},
				"newTopLevel": "newTopValue",
			},
			expected: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2a": map[string]interface{}{
						"key1": "new1",
						"key2": []interface{}{"item3"}, // src slice overwrites dst slice
						"key3": "new3",
					},
					"level2b": map[string]interface{}{
						"nestedInLevel2b": "newValue",
					},
				},
				"level1b": []interface{}{
					map[string]interface{}{
						"arrayKey1": "arrayValue1Modified",
					},
					"arrayItem3", // src slice overwrites dst slice
				},
				"newTopLevel": "newTopValue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeValues(tt.dst, tt.src)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeValuesRecursively() = %v, want %v", result, tt.expected)
			}
		})
	}
}
