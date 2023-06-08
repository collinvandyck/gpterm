package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type tuiState int

const (
	tuiStateInit tuiState = iota
	tuiStateChat
)

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
	return nil
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.state == tuiStateInit {
		m.state = tuiStateChat
		cmds = append(cmds, m.chatModel.Init())
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+h":
			return m, tea.Quit
		}
	}

	switch m.state {
	case tuiStateChat:
		model, cmd := m.chatModel.Update(msg)
		m.chatModel = model.(chatModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	default:
		cmds = append(cmds, tea.Println("unhandled state"))
		cmds = append(cmds, tea.Quit)
	}
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Sequence(cmds...)
}

// View implements tea.Model.
func (m tuiModel) View() string {
	switch m.state {
	case tuiStateChat:
		return m.chatModel.View()
	default:
		return ""
	}
}
