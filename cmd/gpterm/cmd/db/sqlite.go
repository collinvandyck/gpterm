package db

import (
	"os"
	"os/exec"

	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/spf13/cobra"
)

func Sqlite() *cobra.Command {
	return &cobra.Command{
		Use:   "sqlite",
		Short: "open sqlite3 against the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := store.DefaultDBPath()
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
