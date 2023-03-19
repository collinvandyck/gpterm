package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/collinvandyck/gpterm"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/ui"
	"github.com/spf13/cobra"
)

func Repl() *cobra.Command {
	var logfile string
	cmd := &cobra.Command{
		Use:     "repl",
		Short:   "Enter an interactive session",
		Aliases: []string{"repl"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store, err := gpterm.NewStore()
			if err != nil {
				return err
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
			client, err := client.New(key) // todo: add store context
			if err != nil {
				return err
			}
			logger := io.Discard
			if logfile != "" {
				f, err := os.Create(logfile)
				if err != nil {
					return err
				}
				defer f.Close()
				logger = f
			}
			return ui.Start(ctx, store, client, ui.WithLogWriter(logger))
		},
	}
	cmd.Flags().StringVar(&logfile, "log", "", "log to this file")
	return cmd
}
