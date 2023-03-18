package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/collinvandyck/gpterm"
	"github.com/collinvandyck/gpterm/lib/cmdkit"
	"github.com/spf13/cobra"
)

func schemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "updates schema and runs migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpDir, err := os.MkdirTemp("", "")
			if err != nil {
				return fmt.Errorf("mktmp: %w", err)
			}
			defer os.RemoveAll(tmpDir)
			store, err := gpterm.NewStore(gpterm.StoreDir(tmpDir))
			if err != nil {
				return fmt.Errorf("new store: %w", err)
			}
			defer store.Close()
			path := store.DBPath()

			sc := exec.Command("sqlite3", path, ".schema")
			stdout, err := sc.StdoutPipe()
			if err != nil {
				return fmt.Errorf("stdout: %w", err)
			}
			err = sc.Start()
			if err != nil {
				return fmt.Errorf("start: %w", err)
			}
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, stdout)
			if err != nil {
				return fmt.Errorf("copy: %w", err)
			}
			err = sc.Wait()
			if err != nil {
				return fmt.Errorf("wait: %w", err)
			}
			cmdkit.ChdirProject()
			f, err := os.Create("db/schema.sql")
			if err != nil {
				return err
			}
			defer f.Close()
			s := bufio.NewScanner(buf)
			for s.Scan() {
				text := s.Text()
				if !strings.Contains(text, "schema_migrations") {
					fmt.Fprintln(f, text)
				}
			}
			return nil
		},
	}
}
