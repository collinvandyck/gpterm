package cmdkit

import (
	"os"
	"path/filepath"

	"github.com/collinvandyck/gpterm/lib/git"
	"github.com/collinvandyck/gpterm/lib/must"
)

func ChdirProject(paths ...string) {
	pd := must.SucceedVal(git.ProjectDir())
	paths = append([]string{pd}, paths...)
	fullPath := filepath.Join(paths...)
	must.Succeed(os.Chdir(fullPath))
}
