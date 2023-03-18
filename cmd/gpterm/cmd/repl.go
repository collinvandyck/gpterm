package cmd

import (
	"github.com/spf13/cobra"
)

func Repl(tui *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "repl",
		Short: "enter an interactive session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.RunE(tui, nil)
		},
	}
}
