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
	uiOpts
	state     tuiState
	chatModel chatModel
}

func newTUIModel(opts uiOpts) tuiModel {
	return tuiModel{
		uiOpts:    opts.NamedLogger("tui"),
		chatModel: newChatModel(opts),
	}
}

// Init implements tea.Model.
func (m tuiModel) Init() tea.Cmd {
	m.Log("init")
	return nil
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.state == stateInit {
		m.state = stateChat
		cmds = append(cmds, m.chatModel.Init())
	}

	switch m.state {

	case stateChat:
		model, cmd := m.chatModel.Update(msg)
		m.chatModel = model.(chatModel)
		cmds = append(cmds, cmd)
		return m, tea.Sequence(cmds...)

	default:
		cmds = append(cmds, tea.Println("unhandled state"))
		cmds = append(cmds, tea.Quit)
		return m, tea.Sequence(cmds...)
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
