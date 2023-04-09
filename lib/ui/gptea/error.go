package gptea

import tea "github.com/charmbracelet/bubbletea"

type ErrorMsg struct {
	Err error
}

func ErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Err: err}
	}
}
