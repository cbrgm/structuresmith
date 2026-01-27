package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderFileStructureWithPermissions(t *testing.T) {
	// Create a temporary directory for output
	tmpDir, err := os.MkdirTemp("", "structuresmith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := &Structuresmith{
		OutputDir: tmpDir,
	}

	tests := []struct {
		name        string
		file        FileStructure
		wantPerm    os.FileMode
		description string
	}{
		{
			name: "File with executable permissions 0755",
			file: FileStructure{
				Destination: "script.sh",
				Content:     "#!/bin/bash\necho 'hello'",
				Permissions: func() *FileMode { m := FileMode(0o755); return &m }(),
			},
			wantPerm:    0o755,
			description: "Executable script should have 0755 permissions",
		},
		{
			name: "File with default permissions (nil)",
			file: FileStructure{
				Destination: "readme.md",
				Content:     "# README",
				Permissions: nil,
			},
			wantPerm:    0o644,
			description: "File without explicit permissions should default to 0644",
		},
		{
			name: "File with restricted permissions 0600",
			file: FileStructure{
				Destination: "secrets/config.env",
				Content:     "SECRET=value",
				Permissions: func() *FileMode { m := FileMode(0o600); return &m }(),
			},
			wantPerm:    0o600,
			description: "Sensitive file should have 0600 permissions",
		},
		{
			name: "File with read-only permissions 0444",
			file: FileStructure{
				Destination: "readonly.txt",
				Content:     "Do not modify",
				Permissions: func() *FileMode { m := FileMode(0o444); return &m }(),
			},
			wantPerm:    0o444,
			description: "Read-only file should have 0444 permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up output dir for each test
			entries, _ := os.ReadDir(tmpDir)
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}

			err := app.renderFileStructure(tt.file)
			if err != nil {
				t.Fatalf("renderFileStructure() error = %v", err)
			}

			fullPath := filepath.Join(tmpDir, tt.file.Destination)
			info, err := os.Stat(fullPath)
			if err != nil {
				t.Fatalf("Failed to stat created file: %v", err)
			}

			// On Unix, file permissions are ANDed with ^umask, so we check the actual mode
			// We need to mask out the type bits to compare just the permission bits
			gotPerm := info.Mode().Perm()
			if gotPerm != tt.wantPerm {
				t.Errorf("File permissions = %o, want %o (%s)", gotPerm, tt.wantPerm, tt.description)
			}
		})
	}
}

func TestRenderFileStructureWithTemplateAndPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "structuresmith-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := &Structuresmith{
		OutputDir: tmpDir,
	}

	perm := FileMode(0o755)
	file := FileStructure{
		Destination: "templated-script.sh",
		Content:     "#!/bin/bash\necho '{{ .message }}'",
		Values:      map[string]any{"message": "Hello World"},
		Permissions: &perm,
	}

	err = app.renderFileStructure(file)
	if err != nil {
		t.Fatalf("renderFileStructure() error = %v", err)
	}

	fullPath := filepath.Join(tmpDir, file.Destination)

	// Check content was templated correctly
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected := "#!/bin/bash\necho 'Hello World'"
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}

	// Check permissions
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("File permissions = %o, want %o", info.Mode().Perm(), 0o755)
	}
}

func TestProcessDirectoryPreservesPermissions(t *testing.T) {
	// Create a temporary source directory with files
	srcDir, err := os.MkdirTemp("", "structuresmith-src-*")
	if err != nil {
		t.Fatalf("Failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create a test file in source
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := &Structuresmith{}

	// Test that permissions are propagated to processed files
	perm := FileMode(0o755)
	directory := FileStructure{
		Source:      srcDir,
		Destination: "output/",
		Permissions: &perm,
	}

	files, err := app.processDirectory(directory)
	if err != nil {
		t.Fatalf("processDirectory() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	if files[0].Permissions == nil {
		t.Error("Expected permissions to be propagated, got nil")
	} else if *files[0].Permissions != perm {
		t.Errorf("Permissions = %o, want %o", *files[0].Permissions, perm)
	}
}

func TestMergeValuesPreservesPermissions(t *testing.T) {
	// This test verifies that when merging file structures from groups,
	// permissions are handled correctly (they're not in Values, but this
	// ensures the merging logic doesn't interfere with permissions)

	perm := FileMode(0o755)
	file := FileStructure{
		Destination: "script.sh",
		Content:     "#!/bin/bash",
		Values:      map[string]any{"key": "value"},
		Permissions: &perm,
	}

	groupValues := map[string]any{"newKey": "newValue"}
	mergedValues := mergeValues(file.Values, groupValues)

	// Verify values are merged
	if mergedValues["key"] != "value" {
		t.Error("Original value not preserved")
	}
	if mergedValues["newKey"] != "newValue" {
		t.Error("New value not merged")
	}

	// Verify permissions are still on the file (not affected by value merging)
	if file.Permissions == nil || *file.Permissions != perm {
		t.Error("Permissions should not be affected by value merging")
	}
}

func TestShouldOverwrite(t *testing.T) {
	tests := []struct {
		name      string
		overwrite *bool
		want      bool
	}{
		{
			name:      "Overwrite nil (default true)",
			overwrite: nil,
			want:      true,
		},
		{
			name:      "Overwrite explicitly true",
			overwrite: func() *bool { b := true; return &b }(),
			want:      true,
		},
		{
			name:      "Overwrite explicitly false",
			overwrite: func() *bool { b := false; return &b }(),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := FileStructure{
				Destination: "test.txt",
				Content:     "test",
				Overwrite:   tt.overwrite,
			}
			got := shouldOverwrite(file)
			if got != tt.want {
				t.Errorf("shouldOverwrite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplySkipLogic(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "structuresmith-skip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an existing file
	existingFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0o644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	app := &Structuresmith{
		OutputDir: tmpDir,
	}

	overwriteFalse := false
	overwriteTrue := true

	tests := []struct {
		name         string
		diff         DiffResult
		wantNew      int
		wantKept     int
		wantSkipped  int
		wantDeleted  int
		description  string
	}{
		{
			name: "New file with overwrite:false that exists on disk should be skipped",
			diff: DiffResult{
				NewFiles: []FileStructure{
					{Destination: "existing.txt", Content: "new content", Overwrite: &overwriteFalse},
				},
			},
			wantNew:     0,
			wantKept:    0,
			wantSkipped: 1,
			wantDeleted: 0,
			description: "File exists and overwrite is false, should skip",
		},
		{
			name: "New file with overwrite:false that doesn't exist should be created",
			diff: DiffResult{
				NewFiles: []FileStructure{
					{Destination: "new-file.txt", Content: "new content", Overwrite: &overwriteFalse},
				},
			},
			wantNew:     1,
			wantKept:    0,
			wantSkipped: 0,
			wantDeleted: 0,
			description: "File doesn't exist, should create even with overwrite:false",
		},
		{
			name: "Kept file with overwrite:false should be skipped",
			diff: DiffResult{
				KeptFiles: []FileStructure{
					{Destination: "existing.txt", Content: "updated content", Overwrite: &overwriteFalse},
				},
			},
			wantNew:     0,
			wantKept:    0,
			wantSkipped: 1,
			wantDeleted: 0,
			description: "File exists and overwrite is false, should skip overwrite",
		},
		{
			name: "Kept file with overwrite:true should be overwritten",
			diff: DiffResult{
				KeptFiles: []FileStructure{
					{Destination: "existing.txt", Content: "updated content", Overwrite: &overwriteTrue},
				},
			},
			wantNew:     0,
			wantKept:    1,
			wantSkipped: 0,
			wantDeleted: 0,
			description: "File exists but overwrite is true, should overwrite",
		},
		{
			name: "Kept file with overwrite:nil (default) should be overwritten",
			diff: DiffResult{
				KeptFiles: []FileStructure{
					{Destination: "existing.txt", Content: "updated content", Overwrite: nil},
				},
			},
			wantNew:     0,
			wantKept:    1,
			wantSkipped: 0,
			wantDeleted: 0,
			description: "File exists and overwrite is nil (default true), should overwrite",
		},
		{
			name: "Mixed scenario",
			diff: DiffResult{
				NewFiles: []FileStructure{
					{Destination: "brand-new.txt", Content: "new", Overwrite: nil},
					{Destination: "existing.txt", Content: "new", Overwrite: &overwriteFalse},
				},
				KeptFiles: []FileStructure{
					// Note: another-existing.txt doesn't exist on disk, so it stays in KeptFiles
					// (only files that exist on disk AND have overwrite:false are skipped)
					{Destination: "another-not-on-disk.txt", Content: "update", Overwrite: &overwriteFalse},
				},
				DeletedFiles: []FileStructure{
					{Destination: "to-delete.txt"},
				},
			},
			wantNew:     1, // brand-new.txt
			wantKept:    1, // another-not-on-disk.txt stays in kept (doesn't exist on disk)
			wantSkipped: 1, // existing.txt (exists on disk + overwrite:false)
			wantDeleted: 1, // to-delete.txt
			description: "Mixed scenario with various overwrite settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.applySkipLogic(tt.diff)

			if len(result.NewFiles) != tt.wantNew {
				t.Errorf("NewFiles count = %d, want %d (%s)", len(result.NewFiles), tt.wantNew, tt.description)
			}
			if len(result.KeptFiles) != tt.wantKept {
				t.Errorf("KeptFiles count = %d, want %d (%s)", len(result.KeptFiles), tt.wantKept, tt.description)
			}
			if len(result.SkippedFiles) != tt.wantSkipped {
				t.Errorf("SkippedFiles count = %d, want %d (%s)", len(result.SkippedFiles), tt.wantSkipped, tt.description)
			}
			if len(result.DeletedFiles) != tt.wantDeleted {
				t.Errorf("DeletedFiles count = %d, want %d (%s)", len(result.DeletedFiles), tt.wantDeleted, tt.description)
			}
		})
	}
}

func TestRenderWithOverwriteFalse(t *testing.T) {
	// Create a temporary directory for output
	tmpDir, err := os.MkdirTemp("", "structuresmith-overwrite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an existing file with specific content
	existingFile := filepath.Join(tmpDir, "config.env")
	originalContent := "ORIGINAL_SECRET=keep-me"
	if err := os.WriteFile(existingFile, []byte(originalContent), 0o644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	app := &Structuresmith{
		OutputDir: tmpDir,
	}

	overwriteFalse := false

	// Try to render a file with overwrite: false
	file := FileStructure{
		Destination: "config.env",
		Content:     "NEW_SECRET=overwrite-me",
		Overwrite:   &overwriteFalse,
	}

	// First, simulate what would happen - the file exists and has overwrite:false
	// so it should be skipped
	diff := DiffResult{
		KeptFiles: []FileStructure{file},
	}
	result := app.applySkipLogic(diff)

	if len(result.SkippedFiles) != 1 {
		t.Errorf("Expected 1 skipped file, got %d", len(result.SkippedFiles))
	}
	if len(result.KeptFiles) != 0 {
		t.Errorf("Expected 0 kept files, got %d", len(result.KeptFiles))
	}

	// Verify the original file content is preserved
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("File content was modified, got %q, want %q", string(content), originalContent)
	}
}

func TestProcessDirectoryPreservesOverwrite(t *testing.T) {
	// Create a temporary source directory with files
	srcDir, err := os.MkdirTemp("", "structuresmith-src-*")
	if err != nil {
		t.Fatalf("Failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create a test file in source
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	app := &Structuresmith{}

	// Test that overwrite setting is propagated to processed files
	overwriteFalse := false
	directory := FileStructure{
		Source:      srcDir,
		Destination: "output/",
		Overwrite:   &overwriteFalse,
	}

	files, err := app.processDirectory(directory)
	if err != nil {
		t.Fatalf("processDirectory() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	if files[0].Overwrite == nil {
		t.Error("Expected overwrite to be propagated, got nil")
	} else if *files[0].Overwrite != overwriteFalse {
		t.Errorf("Overwrite = %v, want %v", *files[0].Overwrite, overwriteFalse)
	}
}
