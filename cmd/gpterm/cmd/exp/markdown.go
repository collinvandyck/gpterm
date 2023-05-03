package exp

import (
	"bytes"
	"embed"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/collinvandyck/gpterm/lib/markdown"
	"github.com/spf13/cobra"
)

//go:embed markdowns/*
var markdowns embed.FS

func markdownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "markdown",
		Short: "test out markdown renderer",
		RunE: func(cc *cobra.Command, args []string) error {
			cmd := exec.Command("stty", "size")
			cmd.Stdin = os.Stdin
			out, err := cmd.Output()
			if err != nil {
				return err
			}
			ws := strings.Split(string(out), " ")[1]
			width, err := strconv.Atoi(strings.TrimSpace(ws))
			if err != nil {
				return err
			}

			file := "sample-1.md"
			if len(args) > 0 {
				file = args[0]
			}
			bs, err := markdowns.ReadFile("markdowns/" + file)
			if err != nil {
				return err
			}
			bs, err = markdown.RenderBytes(bs, width)
			if err != nil {
				return err
			}
			io.Copy(os.Stdout, bytes.NewReader(bs))
			return nil
		},
	}
}
