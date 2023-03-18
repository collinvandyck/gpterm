package cmdkit

import (
	"os"
	"path/filepath"

	"github.com/collinvandyck/gpterm/lib/git"
	"github.com/collinvandyck/gpterm/lib/must"
)

func ChdirProject(paths ...string) {
	pd := git.MustProjectDir()
	paths = append([]string{pd}, paths...)
	fullPath := filepath.Join(paths...)
	must.Succeed(os.Chdir(fullPath))
}
