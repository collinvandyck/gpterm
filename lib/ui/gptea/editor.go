package gptea

import tea "github.com/charmbracelet/bubbletea"

func StartEditorCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return EditorRequestMsg{Prompt: msg}
	}
}

type EditorRequestMsg struct {
	Prompt string
}

type EditorResultMsg struct {
	Text string
	Err  error
}
