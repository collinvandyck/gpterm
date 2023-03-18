package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/collinvandyck/gpterm"
	"github.com/spf13/cobra"
)

func Repl() *cobra.Command {
	return &cobra.Command{
		Use:   "repl",
		Short: "enter an interactive session",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			store, err := gpterm.NewStore()
			if err != nil {
				return err
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
			gpt, err := gpterm.NewClient(ctx, store)
			if err != nil {
				return err
			}
			defer gpt.Close()
			in := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("> ")
				if !in.Scan() {
					break
				}
				text := in.Text()
				messages, err := gpt.Complete(ctx, text)
				if err != nil {
					log.Fatal(err)
				}
				for _, message := range messages {
					fmt.Println()
					fmt.Println(message)
					fmt.Println()
				}
			}
			return nil
		},
	}
}
