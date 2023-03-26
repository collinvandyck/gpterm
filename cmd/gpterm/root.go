package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd"
	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd/db"
	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd/exp"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/collinvandyck/gpterm/lib/ui"
	"github.com/spf13/cobra"
)

var logfile string
var root = &cobra.Command{
	Use:          filepath.Base(os.Args[0]),
	Short:        "Start an interactive session with ChatGPT",
	Long:         "Long",
	Example:      "Example",
	SilenceUsage: true,
	Aliases:      []string{"repl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		logger := log.Discard
		if logfile != "" {
			f, err := os.Create(logfile)
			if err != nil {
				return err
			}
			defer f.Close()
			logger = log.New(log.WithStdout(f), log.WithStderr(f))
		}
		store, err := store.New(store.StoreLog(log.Prefixed("store", logger)))
		if err != nil {
			return fmt.Errorf("new store: %w", err)
		}
		key, err := store.GetAPIKey(ctx)
		if err != nil {
			return err
		}
		if key == "" {
			fmt.Fprintln(os.Stderr, "No API key has been set. Run this command to set it:")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s auth", cmd.Root().Use))
			os.Exit(1)
		}
		client, err := client.New(key)
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}
		ui := ui.New(store, client, ui.WithLogger(logger))
		return ui.Run(ctx)
	},
}

func init() {
	root.Flags().StringVar(&logfile, "log", "", "log to this file")

	root.AddCommand(cmd.Auth())
	root.AddCommand(cmd.Deps())
	root.AddCommand(cmd.Usage())
	root.AddCommand(db.DB(cmd.Deps()))
	root.AddCommand(exp.Exp(cmd.Deps()))
}

func main() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
