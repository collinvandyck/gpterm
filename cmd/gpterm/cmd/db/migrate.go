package db

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/collinvandyck/gpterm/db"
	"github.com/collinvandyck/gpterm/lib/git"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

func Migrate(deps *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run db migrations",
	}
	cmdNew := &cobra.Command{
		Use:     "new [name]",
		Short:   "Create a new migration",
		Args:    cobra.ExactArgs(1),
		PreRunE: deps.RunE,
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
			dbPath, err := store.DefaultDBPath()
			if err != nil {
				return err
			}
			sourceDriver, err := iofs.New(db.FSMigrations, "migrations")
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
			dbPath, err := store.DefaultDBPath()
			if err != nil {
				return err
			}
			sourceDriver, err := iofs.New(db.FSMigrations, "migrations")
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
	cmdStep := &cobra.Command{
		Use:   "step [n]",
		Short: "migrate up or down (+n | -n)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}
			dbPath, err := store.DefaultDBPath()
			if err != nil {
				return err
			}
			sourceDriver, err := iofs.New(db.FSMigrations, "migrations")
			if err != nil {
				return err
			}
			path := "sqlite3://" + dbPath
			mg, err := migrate.NewWithSourceInstance("iofs", sourceDriver, path)
			if err != nil {
				return err
			}
			fmt.Printf("Stepping %d\n", n)
			err = mg.Steps(int(n))
			switch {
			case errors.Is(err, migrate.ErrNoChange):
			case err != nil:
				return fmt.Errorf("step: %w", err)
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
	cmd.AddCommand(cmdStep)
	return cmd
}
