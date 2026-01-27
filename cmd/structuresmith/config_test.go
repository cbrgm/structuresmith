package main

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestFileModeUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    FileMode
		wantErr bool
	}{
		{
			name:    "Octal string 0644",
			yaml:    `permissions: "0644"`,
			want:    FileMode(0o644),
			wantErr: false,
		},
		{
			name:    "Octal string 0755",
			yaml:    `permissions: "0755"`,
			want:    FileMode(0o755),
			wantErr: false,
		},
		{
			name:    "Octal string without leading zero",
			yaml:    `permissions: "755"`,
			want:    FileMode(0o755),
			wantErr: false,
		},
		{
			name:    "Octal string 0600",
			yaml:    `permissions: "0600"`,
			want:    FileMode(0o600),
			wantErr: false,
		},
		{
			name:    "Octal string 0777",
			yaml:    `permissions: "0777"`,
			want:    FileMode(0o777),
			wantErr: false,
		},
		{
			name:    "Invalid octal string with 8",
			yaml:    `permissions: "0888"`,
			want:    FileMode(0),
			wantErr: true,
		},
		{
			name:    "Invalid octal string with 9",
			yaml:    `permissions: "0799"`,
			want:    FileMode(0),
			wantErr: true,
		},
		{
			name:    "Invalid non-numeric string",
			yaml:    `permissions: "invalid"`,
			want:    FileMode(0),
			wantErr: true,
		},
		{
			name:    "Integer value (parsed as octal in YAML)",
			yaml:    `permissions: 0644`,
			want:    FileMode(0o644),
			wantErr: false,
		},
		{
			name:    "Integer value 0755",
			yaml:    `permissions: 0755`,
			want:    FileMode(0o755),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Permissions FileMode `yaml:"permissions"`
			}
			err := yaml.Unmarshal([]byte(tt.yaml), &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Permissions != tt.want {
				t.Errorf("UnmarshalYAML() = %o, want %o", result.Permissions, tt.want)
			}
		})
	}
}

func TestFileModeString(t *testing.T) {
	tests := []struct {
		name string
		mode FileMode
		want string
	}{
		{
			name: "Mode 0644",
			mode: FileMode(0o644),
			want: "0644",
		},
		{
			name: "Mode 0755",
			mode: FileMode(0o755),
			want: "0755",
		},
		{
			name: "Mode 0600",
			mode: FileMode(0o600),
			want: "0600",
		},
		{
			name: "Mode 0777",
			mode: FileMode(0o777),
			want: "0777",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStructureWithPermissions(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *FileMode
		wantErr bool
	}{
		{
			name: "File with permissions",
			yaml: `
destination: "script.sh"
source: "script.sh.tmpl"
permissions: "0755"
`,
			want:    func() *FileMode { m := FileMode(0o755); return &m }(),
			wantErr: false,
		},
		{
			name: "File without permissions (should be nil)",
			yaml: `
destination: "readme.md"
source: "readme.md.tmpl"
`,
			want:    nil,
			wantErr: false,
		},
		{
			name: "File with default-style permissions",
			yaml: `
destination: "config.yml"
content: "key: value"
permissions: "0644"
`,
			want:    func() *FileMode { m := FileMode(0o644); return &m }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var file FileStructure
			err := yaml.Unmarshal([]byte(tt.yaml), &file)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && file.Permissions != nil {
				t.Errorf("Expected nil permissions, got %v", file.Permissions)
			}
			if tt.want != nil {
				if file.Permissions == nil {
					t.Errorf("Expected permissions %o, got nil", *tt.want)
				} else if *file.Permissions != *tt.want {
					t.Errorf("Permissions = %o, want %o", *file.Permissions, *tt.want)
				}
			}
		})
	}
}

func TestFileStructureWithOverwrite(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *bool
		wantErr bool
	}{
		{
			name: "File with overwrite true",
			yaml: `
destination: "readme.md"
content: "# README"
overwrite: true
`,
			want:    func() *bool { b := true; return &b }(),
			wantErr: false,
		},
		{
			name: "File with overwrite false",
			yaml: `
destination: "config.env"
content: "SECRET=changeme"
overwrite: false
`,
			want:    func() *bool { b := false; return &b }(),
			wantErr: false,
		},
		{
			name: "File without overwrite (should be nil, defaults to true)",
			yaml: `
destination: "readme.md"
content: "# README"
`,
			want:    nil,
			wantErr: false,
		},
		{
			name: "File with both permissions and overwrite",
			yaml: `
destination: "config.env"
content: "SECRET=changeme"
permissions: "0600"
overwrite: false
`,
			want:    func() *bool { b := false; return &b }(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var file FileStructure
			err := yaml.Unmarshal([]byte(tt.yaml), &file)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && file.Overwrite != nil {
				t.Errorf("Expected nil overwrite, got %v", *file.Overwrite)
			}
			if tt.want != nil {
				if file.Overwrite == nil {
					t.Errorf("Expected overwrite %v, got nil", *tt.want)
				} else if *file.Overwrite != *tt.want {
					t.Errorf("Overwrite = %v, want %v", *file.Overwrite, *tt.want)
				}
			}
		})
	}
}
