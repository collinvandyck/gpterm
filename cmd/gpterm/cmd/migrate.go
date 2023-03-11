package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/collinvandyck/gpterm"
	"github.com/collinvandyck/gpterm/lib/git"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run db migrations",
	}
	cmdNew := &cobra.Command{
		Use:     "new [name]",
		Short:   "Create a new migration",
		Args:    cobra.ExactArgs(1),
		PreRunE: Deps().RunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// migrate create -ext sql -dir migrations -seq credential
			dir := git.MustProjectDir()
			err := os.Chdir(filepath.Join(dir, "db"))
			if err != nil {
				return err
			}
			ec := exec.Command("migrate", "create", "-ext", "sql", "-dir", "migrations", "-seq", args[0])
			out, err := ec.CombinedOutput()
			if err != nil {
				fmt.Fprintln(os.Stderr, string(out))
				return err
			}
			return nil
		},
	}
	cmdUp := &cobra.Command{
		Use:   "up",
		Short: "bring migrations up to date",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := gpterm.DefaultDBPath()
			if err != nil {
				return err
			}
			sourceDriver, err := iofs.New(gpterm.FSMigrations, "db/migrations")
			if err != nil {
				return err
			}
			path := "sqlite3://" + dbPath
			mg, err := migrate.NewWithSourceInstance("iofs", sourceDriver, path)
			if err != nil {
				return err
			}
			fmt.Println("Migrating up")
			err = mg.Up()
			switch {
			case errors.Is(err, migrate.ErrNoChange):
			case err != nil:
				return fmt.Errorf("up: %w", err)
			}
			return nil
		},
	}
	cmdDown := &cobra.Command{
		Use:   "down",
		Short: "undo all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := gpterm.DefaultDBPath()
			if err != nil {
				return err
			}
			sourceDriver, err := iofs.New(gpterm.FSMigrations, "db/migrations")
			if err != nil {
				return err
			}
			path := "sqlite3://" + dbPath
			mg, err := migrate.NewWithSourceInstance("iofs", sourceDriver, path)
			if err != nil {
				return err
			}
			fmt.Println("Migrating down")
			err = mg.Down()
			switch {
			case errors.Is(err, migrate.ErrNoChange):
			case err != nil:
				return fmt.Errorf("up: %w", err)
			}
			return nil
		},
	}
	cmdReset := &cobra.Command{
		Use:   "reset",
		Short: "Resets the database by reapplying all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Are you sure you want to do this? [y/n]: ")
			s := bufio.NewScanner(os.Stdin)
			s.Scan()
			choice := s.Text()
			if choice != "y" {
				return nil
			}
			err := cmdDown.RunE(cmdDown, nil)
			if err != nil {
				return fmt.Errorf("down: %w", err)
			}
			err = cmdUp.RunE(cmdUp, nil)
			if err != nil {
				return fmt.Errorf("up: %w", err)
			}
			return nil
		},
	}
	cmd.AddCommand(cmdNew)
	cmd.AddCommand(cmdUp)
	cmd.AddCommand(cmdDown)
	cmd.AddCommand(cmdReset)
	return cmd
}
