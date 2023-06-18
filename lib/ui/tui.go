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
	var cmd tea.Cmd
	switch m.state {
	case tuiStateChat:
		m.chat, cmd = m.chat.Update(msg)
	case tuiStateOptions:
		m.options, cmd = m.options.Update(msg)
	default:
		return tea.Sequence(tea.Println("unknown state"), tea.Quit)
	}
	return cmd
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
	var cmds commands

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

	var cmd tea.Cmd
	switch m.state {
	case tuiStateChat:
		m.chat, cmd = m.chat.Update(msg)
		cmds.Add(cmd)
	case tuiStateOptions:
		m.options, cmd = m.options.Update(msg)
		cmds.Add(cmd)
	default:
		return m, tea.Sequence(
			tea.Println("unhandled state"),
			tea.Quit,
		)
	}
	return m, cmds.Sequence()
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
