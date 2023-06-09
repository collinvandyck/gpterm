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
	current    tea.Model
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

func (m *tuiModel) propagateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case tuiStateChat:
		return m.chat.Update(msg)
	case tuiStateOptions:
		return m.options.Update(msg)
	}
	return m, nil
}

func (m *tuiModel) setState(state tuiState) {
	m.state = state
	switch state {
	case tuiStateChat:
		m.current = m.chat
	case tuiStateOptions:
		m.current = m.options
	}
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		// we can init once we know the size.
		if m.state == tuiStateInit {
			startInChat := true
			if startInChat {
				m.setState(tuiStateChat)
			} else {
				m.setState(tuiStateOptions)
			}
			m.current.Update(msg)
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
	var cmds []tea.Cmd
	switch m.state {
	case tuiStateChat:
		m.state = tuiStateOptions
		model, cmd := m.options.Update(m.windowSize)
		m.options = model.(optionsModel)
		cmds = append(cmds, m.options.Init(), cmd)
	case tuiStateOptions:
		m.state = tuiStateChat
		model, cmd := m.chat.Update(m.windowSize)
		m.chat = model.(chatModel)
		cmds = append(cmds, m.chat.Init(), cmd)
	default:
		return m, nil
	}
	return m, tea.Sequence(cmds...)
}
