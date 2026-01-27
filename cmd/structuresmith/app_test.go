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
