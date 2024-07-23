package mover

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"file-mover/internal/utils"

	"github.com/pterm/pterm"
)

type FileMover struct {
	RootDir  string
	Progress *pterm.ProgressbarPrinter
}

func NewFileMover(rootDir string) *FileMover {
	return &FileMover{
		RootDir: rootDir,
	}
}

func (fm *FileMover) deleteSubDirs() error {
	time.Sleep(500 * time.Millisecond) // Add a small delay before deleting directories

	dirsToDelete := make([]string, 0)

	err := filepath.Walk(fm.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != fm.RootDir {
			dirsToDelete = append(dirsToDelete, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Sort directories by depth (deepest first)
	utils.SortDirsByDepth(dirsToDelete)

	for _, dir := range dirsToDelete {
		err := os.RemoveAll(dir)
		if err != nil {
			return fmt.Errorf("error deleting directory %s: %w", dir, err)
		}
	}

	return nil
}

func (fm *FileMover) MoveFiles() error {
	totalFiles, err := utils.CountFilesToMove(fm.RootDir)
	if err != nil {
		return fmt.Errorf("error counting files: %w", err)
	}

	fm.Progress, _ = pterm.DefaultProgressbar.WithTotal(totalFiles).Start()

	err = fm.traverseAndMove()
	if err != nil {
		return fmt.Errorf("error moving files: %w", err)
	}

	err = fm.deleteSubDirs()
	if err != nil {
		return fmt.Errorf("error deleting subdirectories: %w", err)
	}

	fm.Progress.Stop()
	pterm.Success.Println("File moving completed successfully!")
	return nil
}

func (fm *FileMover) traverseAndMove() error {
	return filepath.Walk(fm.RootDir, func(path string, info utils.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Dir(path) != fm.RootDir {
			destPath := filepath.Join(fm.RootDir, filepath.Base(path))
			err := utils.MoveFile(path, destPath)
			if err != nil {
				return err
			}
			fm.Progress.Increment()
		}
		return nil
	})
}
