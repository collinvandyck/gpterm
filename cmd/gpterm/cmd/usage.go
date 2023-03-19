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

			// gpt-3.5-turbo cost is $0.002 / 1K tokens
			cost := 0.002 * (float64(usage.TotalTokens) / 1000)

			fmt.Printf("Prompt:     %d\n", usage.PromptTokens)
			fmt.Printf("Completion: %d\n", usage.CompletionTokens)
			fmt.Printf("Total:      %d\n", usage.TotalTokens)
			fmt.Printf("Cost:       $%0.02f ($0.002 per 1K tokens)\n", cost)

			return nil
		},
	}
}
