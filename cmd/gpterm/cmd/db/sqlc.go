package db

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/collinvandyck/gpterm/lib/cmdkit"
	"github.com/spf13/cobra"
)

func SqlC(deps *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "sqlc",
		Short: "Generates sqlc queries",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			err := deps.RunE(cmd, nil)
			if err != nil {
				return err
			}
			err = Schema().RunE(cmd, nil)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			cmdkit.ChdirProject("db")
			ec := exec.Command(
				"sqlc",
				"generate",
				"--file",
				"sqlc.yaml")
			out, err := ec.CombinedOutput()
			if err != nil {
				fmt.Fprintln(os.Stderr, string(out))
				return err
			}
			return nil
		},
	}
}
