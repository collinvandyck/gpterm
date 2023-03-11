package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/collinvandyck/gpterm"
	"github.com/spf13/cobra"
)

func Auth() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Sets the OpenAPI key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Print("OpenAPI key: ")
			s := bufio.NewScanner(os.Stdin)
			s.Scan()
			key := s.Text()
			if key == "" {
				return errors.New("No key supplied")
			}
			store, err := gpterm.NewStore()
			if err != nil {
				return fmt.Errorf("store: %w", err)
			}
			err = store.SetAPIKey(ctx, key)
			if err != nil {
				return fmt.Errorf("set api key: %w", err)
			}
			fmt.Println("API key set")
			return nil
		},
	}
}
