package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileInfo = os.FileInfo

func CountFilesToMove(rootDir string) (int, error) {
	count := 0
	err := filepath.Walk(rootDir, func(path string, info FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Dir(path) != rootDir {
			count++
		}
		return nil
	})
	return count, err
}

func MoveFile(src, dst string) error {
	maxRetries := 5
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := moveFileAttempt(src, dst)
		if err == nil {
			return nil // Success
		}

		if os.IsExist(err) {
			// If the file already exists, rename it
			dst = getUniqueFilename(dst)
			continue
		}

		if os.IsPermission(err) {
			// If it's a permission error, don't retry
			return err
		}

		// Calculate delay with exponential backoff
		delay := baseDelay * time.Duration(1<<uint(attempt))
		time.Sleep(delay)
	}

	return fmt.Errorf("failed to move file after %d attempts", maxRetries)
}

func getUniqueFilename(filepath string) string {
	dir, file := path.Split(filepath)
	ext := path.Ext(file)
	name := file[:len(file)-len(ext)]
	counter := 1

	for {
		newPath := path.Join(dir, fmt.Sprintf("%s_%d%s", name, counter, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

func moveFileAttempt(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	err = sourceFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close source file: %w", err)
	}

	err = destFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	err = os.Remove(src)
	if err != nil {
		// If we can't remove the source, try to remove the destination to avoid duplication
		os.Remove(dst)
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

func SortDirsByDepth(dirs []string) {
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], string(os.PathSeparator)) > strings.Count(dirs[j], string(os.PathSeparator))
	})
}

func RemoveEmptyDir(dir string) error {
	return os.Remove(dir)
}