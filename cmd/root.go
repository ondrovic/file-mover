package cmd

import (
	"file-mover/internal/mover"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "file-mover [root directory]",
	Short: "File Mover CLI",
	Long:  "A CLI tool to move files from subdirectories to the root directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir := args[0]
		fm := mover.NewFileMover(rootDir)
		return fm.MoveFiles()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
