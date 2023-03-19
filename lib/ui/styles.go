package ui

import "github.com/charmbracelet/lipgloss"

type styles interface {
	Role(sender string) string
}

type staticStyles struct {
	senders      map[string]lipgloss.Style
	names        map[string]string
	defaultStyle lipgloss.Style
}

func newStaticStyles() staticStyles {
	return staticStyles{
		senders: map[string]lipgloss.Style{
			"user":      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Underline(true),
			"assistant": lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Underline(true),
		},
		names: map[string]string{
			"user":      "You",
			"assistant": "ChatGPT",
		},
		defaultStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")).Underline(true),
	}
}

func (ss staticStyles) Role(role string) string {
	style, ok := ss.senders[role]
	if !ok {
		style = ss.defaultStyle
	}
	name, ok := ss.names[role]
	if !ok {
		name = role
	}
	return style.Render(name)
}
