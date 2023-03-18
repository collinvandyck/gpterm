package main

import (
	"os"
	"path/filepath"

	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd"
	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd/db"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:          filepath.Base(os.Args[0]),
	Short:        "gpterm is a CLI that integrates with OpenAI",
	SilenceUsage: true,
}

func init() {
	root.AddCommand(cmd.Auth())
	root.AddCommand(cmd.Repl())
	root.AddCommand(db.DB(cmd.Deps()))
	root.AddCommand(cmd.Deps())
	root.AddCommand(cmd.Usage())
}

func main() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
