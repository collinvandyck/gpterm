package exp

import (
	"context"
	"fmt"
	"os"

	"github.com/collinvandyck/gpterm"
	"github.com/spf13/cobra"
)

func Tea() *cobra.Command {
	return &cobra.Command{
		Use:   "tea",
		Short: "Run experimental TUI",
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
			gpt, err := gpterm.NewClient(ctx, store)
			if err != nil {
				return err
			}
			defer gpt.Close()
			tea := tea{
				store:  store,
				client: gpt,
			}
			return tea.Run(ctx)
		},
	}
}

type tea struct {
	store  *gpterm.Store
	client *gpterm.Client
}

func (t *tea) Run(ctx context.Context) error {
	return nil
}
