package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/muesli/reflow/wordwrap"
)

type render struct {
	width  int
	styles styles
}

func newRender(styles styles) render {
	return render{
		styles: styles,
	}
}

func (r *render) setWidth(width int) {
	r.width = width
}

func (r render) renderEntry(entry entry, spinner tea.Model) string {
	role := entry.Role
	if entry.err != nil {
		role = "error"
	}
	role = r.styles.Role(entry.Role)
	if entry.spin {
		role += " " + spinner.View()
	}
	content := r.renderContent(entry)
	return strings.Join([]string{role, content}, "\n")
}

func (r render) renderContent(entry entry) string {
	if entry.err != nil {
		res := fmt.Sprintf("*%v*", entry.err.Error())
		return wordwrap.String(res, r.width)
	}
	line := entry.Message.Content
	bs := markdown.Render(line, r.width, 0)
	line = string(bs)
	return line
}
