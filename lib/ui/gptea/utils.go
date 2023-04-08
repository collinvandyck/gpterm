package gptea

import tea "github.com/charmbracelet/bubbletea"

// MessageCmd is a helper function that wraps a message in a tea.Cmd
func MessageCmd[T any](msg T) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
