package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
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
			store, err := store.New()
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
			client, err := client.New(key, client.WithChatContext(store.ChatContext()))
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			logger := log.Discard
			if logfile != "" {
				f, err := os.Create(logfile)
				if err != nil {
					return err
				}
				defer f.Close()
				logger = log.New(log.WithStdout(f), log.WithStderr(f))
			}
			ui := ui.New(store, client, ui.WithLogger(logger))
			return ui.Run(ctx)
		},
	}
	cmd.Flags().StringVar(&logfile, "log", "", "log to this file")
	return cmd
}
