package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"net/http"
	_ "net/http/pprof"

	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd"
	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd/db"
	"github.com/collinvandyck/gpterm/cmd/gpterm/cmd/exp"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/collinvandyck/gpterm/lib/ui"
	"github.com/spf13/cobra"
)

var (
	logfile        string
	requestLogfile string
	pprof          bool
	clientContext  int
)

var root = &cobra.Command{
	Use:          filepath.Base(os.Args[0]),
	Short:        "Start an interactive session with ChatGPT",
	Long:         "Long",
	Example:      "Example",
	SilenceUsage: true,
	Aliases:      []string{"repl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		lw, err := log.FileWriter(logfile)
		if err != nil {
			return err
		}
		defer lw.Close()
		logger := log.New(log.WithWriter(lw))

		rw, err := log.FileWriter(requestLogfile)
		if err != nil {
			return err
		}
		defer rw.Close()
		requestLogger := log.New(log.WithWriter(rw))

		store, err := store.New(store.StoreLog(log.Prefixed("store", logger)))
		if err != nil {
			return fmt.Errorf("new store: %w", err)
		}
		key, err := store.GetAPIKey(ctx)
		if err != nil {
			return err
		}
		if key == "" {
			fmt.Fprintln(os.Stderr, "No API key has been set. Run this command to set it:")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s auth", cmd.Root().Use))
			os.Exit(1)
		}
		client, err := client.New(key, client.WithRequestLogger(requestLogger))
		if err != nil {
			return fmt.Errorf("new client: %w", err)
		}
		ui := ui.New(store, client, ui.WithLogger(logger))
		return ui.Run(ctx)
	},
}

func init() {
	root.Flags().StringVar(&logfile, "log", "", "log to this file")
	root.Flags().StringVar(&requestLogfile, "request-log", "", "log requests to this file")
	root.Flags().BoolVar(&pprof, "pprof", false, "start pprof http server in background")
	root.Flags().IntVarP(&clientContext, "context-size", "c", 5, "number of messages to send as context")

	root.AddCommand(cmd.Auth())
	root.AddCommand(cmd.Deps())
	root.AddCommand(cmd.Usage())
	root.AddCommand(db.DB(cmd.Deps()))
	root.AddCommand(exp.Exp(cmd.Deps()))
}

func main() {
	if pprof {
		go func() {
			address := "localhost:6060"
			err := http.ListenAndServe(address, nil)
			if err != nil {
				log.Error("Failed to start pprof HTTP server: %v", err)
				os.Exit(1)
			}
		}()
	}
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
