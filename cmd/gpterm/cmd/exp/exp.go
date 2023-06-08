package exp

import "github.com/spf13/cobra"

func Exp(deps *cobra.Command) *cobra.Command {
	exp := &cobra.Command{
		Use:   "exp",
		Short: "Home for experiments",
	}
	exp.AddCommand(scrollCmd())
	exp.AddCommand(scrollbackCmd())
	exp.AddCommand(altScreenCmd())
	exp.AddCommand(markdownCmd())
	exp.AddCommand(lipglossCmd())
	exp.AddCommand(optionsCmd())
	return exp
}
