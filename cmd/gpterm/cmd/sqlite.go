package cmd

import (
	"os"
	"os/exec"

	"github.com/collinvandyck/gpterm"
	"github.com/spf13/cobra"
)

func sqliteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sqlite",
		Short: "open sqlite3 against the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := gpterm.DefaultDBPath()
			if err != nil {
				return err
			}
			ec := exec.Command("sqlite3", dbPath)
			ec.Stdin = os.Stdin
			ec.Stdout = os.Stdout
			ec.Stderr = os.Stderr
			err = ec.Run()
			if err != nil {
				return err
			}
			return nil
		},
	}
}
