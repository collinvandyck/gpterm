package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type tuiState int

const (
	tuiStateInit tuiState = iota
	tuiStateChat
	tuiStateOptions
)

type tuiModel struct {
	uiOpts
	state      tuiState
	chat       chatModel
	options    optionsModel
	windowSize tea.WindowSizeMsg
}

func newTUIModel(opts uiOpts) tuiModel {
	return tuiModel{
		uiOpts:  opts.NamedLogger("tui"),
		chat:    newChatModel(opts),
		options: newOptionsModel(opts),
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
		cmds = append(cmds, m.chat.Init())
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+h":
			return m.switchModel()
		case "ctrl+c":
			switch m.state {
			case tuiStateOptions:
				return m.switchModel()
			}
		}
	}

	switch m.state {
	case tuiStateChat:
		model, cmd := m.chat.Update(msg)
		m.chat = model.(chatModel)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tuiStateOptions:
		model, cmd := m.options.Update(msg)
		m.options = model.(optionsModel)
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
		return m.chat.View()
	case tuiStateOptions:
		return m.options.View()
	default:
		return ""
	}
}

func (m tuiModel) switchModel() (tuiModel, tea.Cmd) {
	cmds := []tea.Cmd{tea.ClearScreen}
	switch m.state {
	case tuiStateChat:
		m.state = tuiStateOptions
		model, cmd := m.options.Update(m.windowSize)
		m.options = model.(optionsModel)
		cmds = append(cmds, m.options.Init(), cmd)
		return m, tea.Sequence(cmds...)
	case tuiStateOptions:
		m.state = tuiStateChat
		model, cmd := m.chat.Update(m.windowSize)
		m.chat = model.(chatModel)
		cmds = append(cmds, cmd)
		model, cmd = m.chat.reset()
		m.chat = model.(chatModel)
		cmds = append(cmds, cmd)
		return m, tea.Sequence(cmds...)
	default:
		return m, nil
	}
}
