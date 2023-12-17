package main

import (
	"reflect"
	"testing"
)

func TestValidateDuplicateProjectNames(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfigFile
		wantErr bool
	}{
		// No projects
		{name: "Empty projects list", config: ConfigFile{}, wantErr: false},

		// No duplicates
		{name: "No duplicates", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo1"}, {Name: "repo2"}}}, wantErr: false},

		// Single duplicate
		{name: "Single duplicate", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo1"}, {Name: "repo1"}}}, wantErr: true},

		// Multiple duplicates
		{name: "Multiple duplicates", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo1"}, {Name: "repo1"}, {Name: "repo2"}, {Name: "repo2"}}}, wantErr: true},

		// Long list without duplicates
		{name: "Long list without duplicates", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo1"}, {Name: "repo2"}, {Name: "repo3"}, {Name: "repo4"}}}, wantErr: false},

		// Long list with one duplicate
		{name: "Long list with one duplicate", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo1"}, {Name: "repo2"}, {Name: "repo3"}, {Name: "repo1"}}}, wantErr: true},

		// Different case sensitivity
		{name: "Case sensitive duplicates", config: ConfigFile{Projects: []ProjectConfig{{Name: "Repo1"}, {Name: "repo1"}}}, wantErr: false},

		// All projects have the same name
		{name: "All same name", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo"}, {Name: "repo"}, {Name: "repo"}}}, wantErr: true},

		// Names with special characters
		{name: "Special characters", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo!"}, {Name: "repo@"}, {Name: "repo#"}}}, wantErr: false},

		// Names with whitespace
		{name: "Whitespace names", config: ConfigFile{Projects: []ProjectConfig{{Name: "repo "}, {Name: " repo"}, {Name: "repo"}}}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDuplicateProjectNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDuplicateProjectNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDuplicateTemplateGroups(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfigFile
		wantErr bool
	}{
		{
			name:    "No Groups",
			config:  ConfigFile{TemplateGroups: make(map[string][]FileStructure)},
			wantErr: false,
		},
		{
			name: "Single Group",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{"group1": {}},
			},
			wantErr: false,
		},
		{
			name: "Multiple Groups No Duplicates",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {},
					"group2": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Case Sensitive Group Names",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {},
					"Group1": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Whitespace in Group Names",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group 1":  {},
					"group 2":  {},
					" group 1": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Mixed Characters in Group Names",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1":  {},
					"group_1": {},
					"group-1": {},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDuplicateTemplateGroups()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDuplicateTemplateGroups() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProjectGroupReferences(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfigFile
		wantErr bool
	}{
		{
			name: "Valid Group References",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "group1"}}},
				},
				TemplateGroups: map[string][]FileStructure{
					"group1": {},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Group References",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "nonExistentGroup"}}},
				},
				TemplateGroups: map[string][]FileStructure{
					"group1": {},
				},
			},
			wantErr: true,
		},
		{
			name: "Multiple Projects, One with Invalid Reference",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "group1"}}},
					{Name: "project2", Groups: []TemplateGroupRef{{GroupName: "nonExistentGroup"}}},
				},
				TemplateGroups: map[string][]FileStructure{
					"group1": {},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty Projects List",
			config: ConfigFile{
				Projects:       []ProjectConfig{},
				TemplateGroups: map[string][]FileStructure{"group1": {}},
			},
			wantErr: false,
		},
		{
			name: "Empty Template Groups",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "group1"}}},
				},
				TemplateGroups: map[string][]FileStructure{},
			},
			wantErr: true,
		},
		{
			name: "Valid Multiple Groups",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "group1"}, {GroupName: "group2"}}},
				},
				TemplateGroups: map[string][]FileStructure{"group1": {}, "group2": {}},
			},
			wantErr: false,
		},
		{
			name: "One Valid One Invalid Group",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1", Groups: []TemplateGroupRef{{GroupName: "group1"}, {GroupName: "nonExistentGroup"}}},
				},
				TemplateGroups: map[string][]FileStructure{"group1": {}},
			},
			wantErr: true,
		},
		{
			name: "Project Without Group References",
			config: ConfigFile{
				Projects: []ProjectConfig{
					{Name: "project1"},
				},
				TemplateGroups: map[string][]FileStructure{"group1": {}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateProjectGroupReferences()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectGroupReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURLSchemes(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfigFile
		wantErr bool
	}{
		{
			name: "Valid URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "http://example.com"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "invalid-url"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: ""}},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple Groups, One Invalid URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "http://valid.com"}},
					"group2": {{SourceURL: "invalid-url"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Complex Valid URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "https://www.example.com/path?query=param#fragment"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Local File Path",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "file:///path/to/file"}},
				},
			},
			wantErr: false,
		},
		{
			name: "No Scheme URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "www.example.com"}},
				},
			},
			wantErr: true,
		},
		{
			name: "FTP URL",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "ftp://ftp.example.com"}},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple Groups, All Valid URLs",
			config: ConfigFile{
				TemplateGroups: map[string][]FileStructure{
					"group1": {{SourceURL: "http://example.com"}},
					"group2": {{SourceURL: "https://example.org"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateURLSchemes()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURLSchemes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindProject(t *testing.T) {
	tests := []struct {
		name      string
		config    ConfigFile
		project   string
		want      Project
		wantError bool
	}{
		{
			name: "Project Found",
			config: ConfigFile{
				Projects: []ProjectConfig{{Name: "project1"}},
			},
			project:   "project1",
			want:      Project{Name: "project1"},
			wantError: false,
		},
		{
			name: "Project Not Found",
			config: ConfigFile{
				Projects: []ProjectConfig{{Name: "project1"}},
			},
			project:   "project2",
			want:      Project{},
			wantError: true,
		},
		{
			name: "Empty Projects List",
			config: ConfigFile{
				Projects: []ProjectConfig{},
			},
			project:   "project1",
			want:      Project{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.FindProject(tt.project)
			if (err != nil) != tt.wantError {
				t.Errorf("FindProject() error = %v, wantError %v", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindProject() = %v, want %v", got, tt.want)
			}
		})
	}
}
