package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/collinvandyck/gpterm/lib/git"
	"github.com/spf13/cobra"
)

func Deps() *cobra.Command {
	return &cobra.Command{
		Use:   "install-deps",
		Short: "Install dependencies into bin",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := git.MustProjectDir()
			err := os.Setenv("GOBIN", filepath.Join(projectDir, "bin"))
			if err != nil {
				return err
			}
			pkgs := []string{
				"github.com/golang-migrate/migrate/v4/cmd/migrate",
				"github.com/kyleconroy/sqlc/cmd/sqlc",
				"golang.org/x/tools/cmd/stringer",
				"github.com/charmbracelet/glow",
			}
			for _, pkg := range pkgs {
				bin := filepath.Join(projectDir, "bin", filepath.Base(pkg))
				_, err := os.Stat(bin)
				switch {
				case os.IsNotExist(err):
				case err != nil:
					return err
				default:
					continue
				}
				fmt.Printf("Installing %s\n", pkg)
				ec := exec.Command("go", "install", pkg)
				out, err := ec.CombinedOutput()
				if err != nil {
					fmt.Fprintln(os.Stderr, string(out))
					return err
				}
			}
			return nil
		},
	}
}
