package db

import (
	"os"
	"os/exec"
	"strings"

	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/spf13/cobra"
)

func Sqlite() *cobra.Command {
	return &cobra.Command{
		Use:   "sqlite",
		Short: "open sqlite3 against the database",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
			HiddenDefaultCmd:    true,
		},
		DisableFlagParsing:    true,
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		DisableAutoGenTag:     true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := store.DefaultDBPath()
			if err != nil {
				return err
			}
			var ec *exec.Cmd
			if len(args) > 0 {
				log.Info("Executing sqlite3 %s '%s'", dbPath, strings.Join(args, " "))
				ec = exec.Command("sqlite3", dbPath, strings.Join(args, " "))
			} else {
				ec = exec.Command("sqlite3", dbPath)
			}
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
