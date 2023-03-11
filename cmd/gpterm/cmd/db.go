package cmd

import (
	"github.com/spf13/cobra"
)

func DB() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Run db related commands",
	}
	cmd.AddCommand(schemaCmd())
	cmd.AddCommand(sqlcCmd())
	cmd.AddCommand(migrateCmd())
	cmd.AddCommand(sqliteCmd())
	return cmd
}
