package markdown

import (
	"bytes"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// writer uses lipgloss to ensure lines are only of a certain width
type writer struct {
	io.Writer
	width int
	buf   bytes.Buffer
	style lipgloss.Style
}

func newWriter(w io.Writer, width int) *writer {
	width = 50
	return &writer{
		Writer: w,
		width:  width,
		style:  lipgloss.NewStyle().Width(width),
	}
}
func (w *writer) Write(p []byte) (n int, err error) {
	parts := bytes.Split(p, []byte("\n"))
	_ = parts
	for i, part := range parts {
		w.buf.Write(part)
		if i > 0 {
			// the buffer reprsents a newline, so we can write it
			line := w.buf.String()
			strings.TrimRight(line, "\n")
			line = w.style.Render(line)
			w.Writer.Write([]byte(line))
			w.buf.Reset()
		}
	}
	return w.Writer.Write(p)
}
