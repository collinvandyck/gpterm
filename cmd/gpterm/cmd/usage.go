package cmd

import (
	"context"
	"fmt"

	"github.com/collinvandyck/gpterm"
	"github.com/spf13/cobra"
)

func Usage() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Display usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store, err := gpterm.NewStore()
			if err != nil {
				return err
			}
			usage, err := store.GetTotalUsage(ctx)
			if err != nil {
				return err
			}
			fmt.Println("Prompt:    ", usage.PromptTokens)
			fmt.Println("Completion:", usage.CompletionTokens)
			fmt.Println("Total:     ", usage.TotalTokens)
			return nil
		},
	}
}
