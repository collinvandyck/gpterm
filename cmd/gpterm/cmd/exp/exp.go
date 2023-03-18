package exp

import "github.com/spf13/cobra"

func Exp(deps *cobra.Command) *cobra.Command {
	exp := &cobra.Command{
		Use:   "exp",
		Short: "Home for experiments",
	}
	exp.AddCommand(Tea())
	return exp
}
