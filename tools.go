//go:build tools
// +build tools

package gpterm

import (
	_ "github.com/charmbracelet/glow"
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
	_ "golang.org/x/tools/cmd/stringer"
)
