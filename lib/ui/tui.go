package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type tuiState int

const (
	stateInit tuiState = iota
	stateChat
)

var _ tea.Model = tuiModel{}

type tuiModel struct {
	state     tuiState
	chatModel chatModel
}

func newTUIModel(opts uiOpts) tuiModel {
	return tuiModel{
		chatModel: newChatModel(opts),
	}
}

// Init implements tea.Model.
func (m tuiModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {

	case stateInit:
		m.state = stateChat
		cmd = m.chatModel.Init()
		return m, cmd

	case stateChat:
		model, cmd := m.chatModel.Update(msg)
		m.chatModel = model.(chatModel)
		return m, cmd

	default:
		cmd = tea.Println("unhandled state")
		return m, tea.Sequence(cmd, tea.Quit)
	}
}

// View implements tea.Model.
func (m tuiModel) View() string {
	switch m.state {
	case stateChat:
		return m.chatModel.View()
	default:
		return ""
	}
}
