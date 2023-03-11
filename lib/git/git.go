package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func MustProjectDir() string {
	dir, err := ProjectDir()
	if err != nil {
		fmt.Println("No project dir detected")
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
