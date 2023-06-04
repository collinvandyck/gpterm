package ui

import (
	"fmt"

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
	return nil
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("Update", "msg", msg, "type", fmt.Sprintf("%T", msg))
	var cmds []tea.Cmd
	if m.state == stateInit {
		m.state = stateChat
		cmds = append(cmds, m.chatModel.Init())
	}
	switch m.state {
	case stateChat:
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
	m.Log("Sending commands", "cmds", cmds)
	return m, tea.Sequence(cmds...)
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
