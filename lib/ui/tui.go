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

func (s tuiState) String() string {
	switch s {
	case tuiStateInit:
		return "init"
	case tuiStateChat:
		return "chat"
	case tuiStateOptions:
		return "options"
	default:
		return "unknown"
	}
}

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

func (m *tuiModel) currentUpdate(msg tea.Msg) tea.Cmd {
	switch m.state {
	case tuiStateChat:
		model, cmd := m.chat.Update(msg)
		m.chat = model.(chatModel)
		return cmd
	case tuiStateOptions:
		model, cmd := m.options.Update(msg)
		m.options = model.(optionsModel)
		return cmd
	default:
		return tea.Sequence(tea.Println("unknown state"), tea.Quit)
	}
}

func (m *tuiModel) setState(state tuiState) {
	m.Log("Setting state", "state", state)
	m.state = state
	switch state {
	case tuiStateChat:
	case tuiStateOptions:
	}
}

func (m *tuiModel) currentInit() tea.Cmd {
	switch m.state {
	case tuiStateChat:
		return m.chat.Init()
	case tuiStateOptions:
		return m.options.Init()
	default:
		return tea.Sequence(tea.Println("unknown state"), tea.Quit)
	}
}

func (m tuiModel) switchModel() (tuiModel, tea.Cmd) {
	var cmds []tea.Cmd
	switch m.state {
	case tuiStateInit:
		m.setState(tuiStateOptions)
	case tuiStateChat:
		m.setState(tuiStateOptions)
	case tuiStateOptions:
		m.setState(tuiStateChat)
	default:
		return m, nil
	}
	cmds = append(cmds,
		tea.ClearScreen,
		m.currentUpdate(m.windowSize),
		m.currentInit(),
	)
	return m, tea.Sequence(cmds...)
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// important to set the window size beforehand
		m.windowSize = msg
		if m.state == tuiStateInit {
			// set the model if this is our first window size
			return m.switchModel()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+h":
			switch m.state {
			case tuiStateChat:
				if !m.chat.isReady() {
					return m, nil
				}
			}
			return m.switchModel()
		case "ctrl+c":
			switch m.state {
			case tuiStateOptions:
				return m, tea.Quit
				//return m.switchModel()
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
