package ui

import "github.com/charmbracelet/lipgloss"

type styles interface {
	Role(sender string) string
	Name(sender string) string
}

type staticStyles struct {
	senders      map[string]lipgloss.Style
	names        map[string]string
	defaultStyle lipgloss.Style
}

func newStaticStyles() staticStyles {
	senderStyle := func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(color)
	}
	return staticStyles{
		senders: map[string]lipgloss.Style{
			"user":      senderStyle(lipgloss.Color("2")),
			"assistant": senderStyle(lipgloss.Color("4")),
			"error":     senderStyle(lipgloss.Color("#ff0000")),
		},
		names: map[string]string{
			"user":      "You",
			"assistant": "ChatGPT",
			"error":     "Error",
		},
		defaultStyle: senderStyle(lipgloss.Color("3")),
	}
}

func (ss staticStyles) Name(role string) string {
	name, ok := ss.names[role]
	if !ok {
		name = role
	}
	return name
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
