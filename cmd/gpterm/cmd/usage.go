package cmd

import (
	"context"

	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/spf13/cobra"
)

func Usage() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Display usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store, err := store.New()
			if err != nil {
				return err
			}
			usage, err := store.GetTotalUsage(ctx)
			if err != nil {
				return err
			}

			// gpt-3.5-turbo cost is $0.002 / 1K tokens
			cost := 0.002 * (float64(usage.TotalTokens) / 1000)

			log.Info("Prompt:     %d", usage.PromptTokens)
			log.Info("Completion: %d", usage.CompletionTokens)
			log.Info("Total:      %d", usage.TotalTokens)
			log.Info("Cost:       $%0.02f ($0.002 per 1K tokens)", cost)

			return nil
		},
	}
}
