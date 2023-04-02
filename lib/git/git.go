package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/collinvandyck/gpterm/lib/log"
)

func MustProjectDir() string {
	dir, err := ProjectDir()
	if err != nil {
		log.Println("No project dir detected", "err", err)
		os.Exit(1)
	}
	return dir
}

func ProjectDir() (string, error) {
	ec := exec.Command("git", "rev-parse", "--show-toplevel")
	root, err := ec.CombinedOutput()
	if err != nil {
		return "", err
	}
	projectDir := strings.TrimSpace(string(root))
	return projectDir, nil
}
