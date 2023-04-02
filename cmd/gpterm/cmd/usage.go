package cmd

import (
	"context"
	"fmt"

	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/spf13/cobra"
)

func Usage() *cobra.Command {
	return &cobra.Command{
		Use:        "usage",
		Short:      "Display usage",
		Deprecated: "stream responses do not yet include usage data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Deprecated != "" {
				return nil
			}
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

			print := func(val string, args ...any) {
				fmt.Printf(val+"\n", args...)
			}

			print("Prompt:     %d", usage.PromptTokens)
			print("Completion: %d", usage.CompletionTokens)
			print("Total:      %d", usage.TotalTokens)
			print("Cost:       $%0.02f ($0.002 per 1K tokens)", cost)

			return nil
		},
	}
}
