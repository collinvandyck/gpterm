package db

import (
	"github.com/spf13/cobra"
)

func DB(deps *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Run db related commands",
	}
	cmd.AddCommand(Schema())
	cmd.AddCommand(SqlC(deps))
	cmd.AddCommand(Migrate(deps))
	cmd.AddCommand(Sqlite())
	return cmd
}
