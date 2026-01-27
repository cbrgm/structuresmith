package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

// AnvilLock represents the structure of the anvil.lock file.
type AnvilLock struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Version     string               `json:"version"`
	Files       []AnvilLockFileEntry `json:"files"`
}

// AnvilLockFileEntry represents an entry in the lock file.
type AnvilLockFileEntry struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum,omitempty"` // Optional checksum for the file
}

func LoadOrCreateLockFile(dir string) (*AnvilLock, error) {
	lock, err := LoadLockFile(dir)
	if err != nil {
		// If there's an error loading, try to create a new lock file.
		lock, err = NewLockFile(dir)
		if err != nil {
			return nil, fmt.Errorf("error creating new lock file: %v", err)
		}
	}
	return lock, nil
}

func NewLockFile(dir string) (*AnvilLock, error) {
	lock := AnvilLock{
		GeneratedAt: time.Now(),
		Version:     Version, // Ensure Version is defined elsewhere in your package.
		Files:       []AnvilLockFileEntry{},
	}

	err := lock.saveToDisk(dir)
	if err != nil {
		return nil, err
	}
	return &lock, nil
}

// WriteLockFile creates and saves an AnvilLock with the provided file entries to the specified directory.
func WriteLockFile(fileStructures []FileStructure, dir string) error {
	lock := AnvilLock{
		GeneratedAt: time.Now(),
		Version:     Version,
	}
	fileEntries := make([]AnvilLockFileEntry, len(fileStructures))
	for i, fs := range fileStructures {
		fileEntries[i] = lock.convertToFileEntry(fs)
	}
	lock.Files = fileEntries

	return lock.saveToDisk(dir)
}

// saveToDisk saves the AnvilLock to the specified directory.
func (a *AnvilLock) saveToDisk(dir string) error {
	lockFilePath := filepath.Join(dir, ".anvil.lock")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating directory %s: %v", dir, err)
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling anvil.lock data: %v", err)
	}

	return os.WriteFile(lockFilePath, data, 0o644)
}

// LoadLockFile loads the anvil.lock file from a given directory.
func LoadLockFile(dir string) (*AnvilLock, error) {
	lockFilePath := filepath.Join(dir, ".anvil.lock")
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".anvil.lock file does not exist in %s", dir)
	}

	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading .anvil.lock file: %v", err)
	}

	var lock AnvilLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("error parsing .anvil.lock file: %v", err)
	}

	return &lock, nil
}

// convertToFileEntry converts a FileStructure to an AnvilLockFileEntry.
func (a *AnvilLock) convertToFileEntry(fileStructure FileStructure) AnvilLockFileEntry {
	// TODO: Implement checksum calculation.
	return AnvilLockFileEntry{
		Path:     fileStructure.Destination,
		Checksum: "", // Placeholder for checksum.
	}
}

// FileStatus constants representing the status of a file in the diff.
type FileStatus string

const (
	StatusNew     FileStatus = "New"
	StatusDeleted FileStatus = "Deleted"
	StatusKept    FileStatus = "Kept"
	StatusSkipped FileStatus = "Skipped"
)

// DiffResult represents the result of diffing FileStructures against AnvilLock entries.
type DiffResult struct {
	NewFiles     []FileStructure // Files present in FileStructures but not in AnvilLock.
	DeletedFiles []FileStructure // Files present in AnvilLock but not in FileStructures.
	KeptFiles    []FileStructure // Files present in both AnvilLock and FileStructures.
	SkippedFiles []FileStructure // Files that exist on disk and have overwrite: false.
}

func (d DiffResult) String() string {
	fileMap := make(map[string]FileStatus)

	// Populate the map
	for _, file := range d.NewFiles {
		fileMap[file.Destination] = StatusNew
	}
	for _, file := range d.DeletedFiles {
		fileMap[file.Destination] = StatusDeleted
	}
	for _, file := range d.KeptFiles {
		fileMap[file.Destination] = StatusKept
	}
	for _, file := range d.SkippedFiles {
		fileMap[file.Destination] = StatusSkipped
	}

	// Sort the keys (file paths)
	keys := make([]string, 0, len(fileMap))
	for key := range fileMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Initialize tabwriter
	var result strings.Builder
	writer := tabwriter.NewWriter(&result, 0, 8, 2, ' ', 0)

	// Build the string with tabwriter
	for _, key := range keys {
		status := fileMap[key]
		prefix := getColorAndPrefix(status)
		_, _ = fmt.Fprintf(writer, "%s\t%s\n", prefix, key)
	}

	if err := writer.Flush(); err != nil {
		return "Error generating output"
	}
	return result.String()
}

func getColorAndPrefix(status FileStatus) string {
	switch status {
	case StatusNew:
		return color.New(color.FgGreen).Sprintf("new:")
	case StatusDeleted:
		return color.New(color.FgRed).Sprintf("delete:")
	case StatusKept:
		return color.New(color.FgYellow).Sprintf("overwrite:")
	case StatusSkipped:
		return color.New(color.FgCyan).Sprintf("skip:")
	default:
		return "n/a: "
	}
}

// findNewFiles finds FileStructures that are not in AnvilLock.
func (a *AnvilLock) findNewFiles(fileStructures []FileStructure) []FileStructure {
	lockFileSet := make(map[string]struct{})
	for _, entry := range a.Files {
		lockFileSet[entry.Path] = struct{}{}
	}

	var newFiles []FileStructure
	for _, fs := range fileStructures {
		if _, exists := lockFileSet[fs.Destination]; !exists {
			newFiles = append(newFiles, fs)
		}
	}

	return newFiles
}

// findDeletedFiles finds FileStructures that are in AnvilLock but not in the provided fileStructures.
func (a *AnvilLock) findDeletedFiles(fileStructures []FileStructure) []FileStructure {
	fileStructureSet := make(map[string]struct{})
	for _, fs := range fileStructures {
		fileStructureSet[fs.Destination] = struct{}{}
	}

	var deletedFiles []FileStructure
	for _, entry := range a.Files {
		if _, exists := fileStructureSet[entry.Path]; !exists {
			deletedFile := FileStructure{
				Destination: entry.Path,
				// Other fields of FileStructure are unknown for deleted files
			}
			deletedFiles = append(deletedFiles, deletedFile)
		}
	}

	return deletedFiles
}

// findKeptFiles finds FileStructures that are in both the provided fileStructures and AnvilLock.
func (a *AnvilLock) findKeptFiles(fileStructures []FileStructure) []FileStructure {
	lockFileSet := make(map[string]struct{})
	for _, entry := range a.Files {
		lockFileSet[entry.Path] = struct{}{}
	}

	var keptFiles []FileStructure
	for _, fs := range fileStructures {
		if _, exists := lockFileSet[fs.Destination]; exists {
			keptFiles = append(keptFiles, fs)
		}
	}

	return keptFiles
}

// Diff performs a full diff operation, combining new, deleted, and kept files.
func (a *AnvilLock) Diff(fileStructures []FileStructure) DiffResult {
	return DiffResult{
		NewFiles:     a.findNewFiles(fileStructures),
		DeletedFiles: a.findDeletedFiles(fileStructures),
		KeptFiles:    a.findKeptFiles(fileStructures),
	}
}
