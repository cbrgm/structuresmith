package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNewFiles(t *testing.T) {
	tests := []struct {
		name           string
		lock           AnvilLock
		fileStructures []FileStructure
		want           []FileStructure
	}{
		{
			name:           "New Files Found",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "existing.txt"}}},
			fileStructures: []FileStructure{{Destination: "new.txt"}, {Destination: "existing.txt"}},
			want:           []FileStructure{{Destination: "new.txt"}},
		},
		{
			name:           "No New Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "existing.txt"}}},
			fileStructures: []FileStructure{{Destination: "existing.txt"}},
			want:           []FileStructure{},
		},
		{
			name:           "Empty FileStructures",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "existing.txt"}}},
			fileStructures: []FileStructure{},
			want:           []FileStructure{},
		},
		{
			name:           "Multiple New and Existing Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}, {Destination: "new1.txt"}, {Destination: "file2.txt"}, {Destination: "new2.txt"}},
			want:           []FileStructure{{Destination: "new1.txt"}, {Destination: "new2.txt"}},
		},
		{
			name:           "New Files with Duplicates",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}}},
			fileStructures: []FileStructure{{Destination: "new.txt"}, {Destination: "new.txt"}},
			want:           []FileStructure{{Destination: "new.txt"}, {Destination: "new.txt"}},
		},
		{
			name:           "All New Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{}},
			fileStructures: []FileStructure{{Destination: "new1.txt"}, {Destination: "new2.txt"}},
			want:           []FileStructure{{Destination: "new1.txt"}, {Destination: "new2.txt"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lock.findNewFiles(tt.fileStructures)

			if len(got) != len(tt.want) {
				t.Fatalf("findNewFiles() returned %d items, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Destination != tt.want[i].Destination {
					t.Errorf("findNewFiles() item %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindDeletedFiles(t *testing.T) {
	tests := []struct {
		name           string
		lock           AnvilLock
		fileStructures []FileStructure
		want           []FileStructure
	}{
		{
			name:           "Deleted Files Found",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}},
			want:           []FileStructure{{Destination: "file2.txt"}},
		},
		{
			name:           "No Deleted Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}},
			want:           []FileStructure{},
		},
		{
			name:           "All Files Deleted",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{},
			want:           []FileStructure{{Destination: "file1.txt"}, {Destination: "file2.txt"}},
		},
		{
			name:           "Mixed Deleted and Existing Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}, {Path: "file3.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}, {Destination: "file3.txt"}},
			want:           []FileStructure{{Destination: "file2.txt"}},
		},
		{
			name:           "Deleted Files with Duplicates in Lock",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file1.txt"}}},
			fileStructures: []FileStructure{{Destination: "file2.txt"}},
			want:           []FileStructure{{Destination: "file1.txt"}, {Destination: "file1.txt"}},
		},
		{
			name:           "Empty Lock File Entries",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}},
			want:           []FileStructure{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lock.findDeletedFiles(tt.fileStructures)

			if len(got) != len(tt.want) {
				t.Fatalf("findDeletedFiles() returned %d items, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Destination != tt.want[i].Destination {
					t.Errorf("findDeletedFiles() item %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFindKeptFiles(t *testing.T) {
	tests := []struct {
		name           string
		lock           AnvilLock
		fileStructures []FileStructure
		want           []FileStructure
	}{
		{
			name:           "Kept Files Found",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}, {Destination: "new.txt"}},
			want:           []FileStructure{{Destination: "file1.txt"}},
		},
		{
			name:           "No Kept Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}}},
			fileStructures: []FileStructure{{Destination: "new.txt"}},
			want:           []FileStructure{},
		},
		{
			name:           "All Files Kept",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}, {Destination: "file2.txt"}},
			want:           []FileStructure{{Destination: "file1.txt"}, {Destination: "file2.txt"}},
		},
		{
			name:           "Mixed Kept and New Files",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file2.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}, {Destination: "new.txt"}},
			want:           []FileStructure{{Destination: "file1.txt"}},
		},
		{
			name:           "Kept Files with Duplicates",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}, {Path: "file1.txt"}}},
			fileStructures: []FileStructure{{Destination: "file1.txt"}},
			want:           []FileStructure{{Destination: "file1.txt"}},
		},
		{
			name:           "Empty FileStructures",
			lock:           AnvilLock{Files: []AnvilLockFileEntry{{Path: "file1.txt"}}},
			fileStructures: []FileStructure{},
			want:           []FileStructure{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lock.findKeptFiles(tt.fileStructures)

			if len(got) != len(tt.want) {
				t.Fatalf("findKeptFiles() returned %d items, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Destination != tt.want[i].Destination {
					t.Errorf("findKeptFiles() item %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoadLockFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error // Function to set up test scenario
		wantError bool
	}{
		{
			name: "Lock File Exists but is Invalid JSON",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".anvil.lock"), []byte("invalid json"), 0o666)
			},
			wantError: true,
		},
		{
			name: "Lock File Exists but is Empty",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".anvil.lock"), []byte(""), 0o666)
			},
			wantError: true,
		},
		{
			name: "Lock File Exists but is Inaccessible",
			setupFunc: func(dir string) error {
				file := filepath.Join(dir, ".anvil.lock")
				if err := os.WriteFile(file, []byte("{}"), 0o666); err != nil {
					return err
				}
				return os.Chmod(file, 0o000) // No permission to read
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "testLoadLockFile")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir) // Clean up after the test

			if err = tt.setupFunc(dir); err != nil {
				t.Fatal(err)
			}

			_, err = LoadLockFile(dir)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadLockFile() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
