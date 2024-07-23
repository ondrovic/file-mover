package mover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pterm/pterm"
)

func setupTestDirectory(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "filemover_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test directory structure
	dirs := []string{"subdir1", "subdir2", "subdir1/nesteddir"}
	for _, dir := range dirs {
		err = os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
	}

	// Create test files
	files := map[string][]string{
		"":                  {"root1.txt", "root2.txt"},
		"subdir1":           {"file1.txt", "file2.txt"},
		"subdir2":           {"file3.txt"},
		"subdir1/nesteddir": {"file4.txt", "file5.txt"},
	}

	for dir, fileList := range files {
		for _, file := range fileList {
			err = os.WriteFile(filepath.Join(tempDir, dir, file), []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	return tempDir, func() { os.RemoveAll(tempDir) }
}

func TestFileMover_MoveFiles(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	fm := NewFileMover(tempDir) // false for not dry-run
	err := fm.MoveFiles()
	if err != nil {
		t.Fatalf("FileMover.MoveFiles failed: %v", err)
	}

	// Check if all files were moved to root directory
	expectedFiles := []string{"root1.txt", "root2.txt", "file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("File %s was not found in root directory", file)
		}
	}

	// Check if subdirectories were deleted
	subDirs := []string{"subdir1", "subdir2", "subdir1/nesteddir"}
	for _, dir := range subDirs {
		if _, err := os.Stat(filepath.Join(tempDir, dir)); !os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not deleted", dir)
		}
	}
}

func TestFileMover_traverseAndMove(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	fm := NewFileMover(tempDir)
	fm.Progress, _ = pterm.DefaultProgressbar.WithTotal(5).Start() // 5 files to move

	err := fm.traverseAndMove()
	if err != nil {
		t.Fatalf("FileMover.traverseAndMove failed: %v", err)
	}

	// Check if files were moved to root directory
	expectedFiles := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("File %s was not moved to root directory", file)
		}
	}

	// Check if files were removed from subdirectories
	subDirFiles := []string{
		"subdir1/file1.txt",
		"subdir1/file2.txt",
		"subdir2/file3.txt",
		"subdir1/nesteddir/file4.txt",
		"subdir1/nesteddir/file5.txt",
	}
	for _, file := range subDirFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); !os.IsNotExist(err) {
			t.Errorf("File %s still exists in subdirectory", file)
		}
	}

	// Check if progress bar was updated correctly
	if fm.Progress.Total != 5 {
		t.Errorf("Expected progress bar total to be 5, got %d", fm.Progress.Total)
	}
}

func TestFileMover_deleteSubDirs(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	fm := NewFileMover(tempDir)
	err := fm.deleteSubDirs()
	if err != nil {
		t.Fatalf("FileMover.deleteSubDirs failed: %v", err)
	}

	// Check if subdirectories were deleted
	subDirs := []string{"subdir1", "subdir2", "subdir1/nesteddir"}
	for _, dir := range subDirs {
		if _, err := os.Stat(filepath.Join(tempDir, dir)); !os.IsNotExist(err) {
			t.Errorf("Subdirectory %s was not deleted", dir)
		}
	}

	// Check if root directory still exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Root directory was incorrectly deleted")
	}

	// Check if files in root directory still exist
	rootFiles := []string{"root1.txt", "root2.txt"}
	for _, file := range rootFiles {
		if _, err := os.Stat(filepath.Join(tempDir, file)); os.IsNotExist(err) {
			t.Errorf("File %s in root directory was incorrectly deleted", file)
		}
	}
}