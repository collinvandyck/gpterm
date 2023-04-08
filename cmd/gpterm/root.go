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
	clientHistory  int
)

var root = &cobra.Command{
	Use:          filepath.Base(os.Args[0]),
	Short:        "Start an interactive session with ChatGPT",
	SilenceUsage: true,
	Aliases:      []string{"repl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		lw, err := log.FileWriter(logfile)
		if err != nil {
			return err
		}
		defer lw.Close()
		logger := log.NewWriter(lw)

		rw, err := log.FileWriter(requestLogfile)
		if err != nil {
			return err
		}
		defer rw.Close()
		requestLogger := log.NewWriter(rw)

		store, err := store.New(store.StoreLog(logger.New("name", "store")))
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
		ui := ui.New(store, client, ui.WithLogger(logger), ui.WithClientHistory(clientHistory))
		return ui.Run(ctx)
	},
}

func init() {
	root.Flags().StringVar(&logfile, "log", "", "log to this file")
	root.Flags().StringVar(&requestLogfile, "request-log", "", "log HTTP requests to this file")
	root.Flags().BoolVar(&pprof, "pprof", false, "start pprof http server in background")
	root.Flags().IntVarP(&clientHistory, "context-size", "c", 5, "number of messages to send as context")

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
				fmt.Fprintf(os.Stderr, "Failed to start pprof HTTP server: %v\n", err)
				os.Exit(1)
			}
		}()
	}
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
