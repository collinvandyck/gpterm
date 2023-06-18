package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type tuiState int

const (
	tuiStateChat tuiState = iota
	tuiStateOptions
)

type tuiModel struct {
	uiOpts
	chat       model
	options    model
	state      tuiState
	windowSize tea.WindowSizeMsg
	ready      bool
}

func newTUIModel(opts uiOpts) tuiModel {
	res := tuiModel{
		uiOpts:  opts.NamedLogger("tui"),
		chat:    newChatModel(opts),
		options: newOptionsModel(opts),
	}
	res.state = tuiStateOptions
	return res
}

// Init implements tea.Model.
func (m tuiModel) Init() tea.Cmd {
	if m.state == tuiStateChat {
		return m.chat.Init()
	}
	return m.options.Init()
}

func (m tuiModel) switchModel() (tuiModel, tea.Cmd) {
	switch m.state {
	case tuiStateChat:
		m.state = tuiStateOptions
	case tuiStateOptions:
		m.state = tuiStateChat
	}
	var (
		updateCmd tea.Cmd
		initCmd   tea.Cmd
	)
	switch m.state {
	case tuiStateChat:
		m.chat, updateCmd = m.chat.Update(m.windowSize)
		initCmd = m.chat.Init()
	case tuiStateOptions:
		m.options, updateCmd = m.options.Update(m.windowSize)
		initCmd = m.options.Init()
	}
	return m, tea.Sequence(
		tea.ClearScreen,
		updateCmd,
		initCmd,
	)
}

// Update implements tea.Model.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds commands
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// we must pass this message to all models
		var (
			chatCmd    tea.Cmd
			optionsCmd tea.Cmd
		)
		m.chat, chatCmd = m.chat.Update(msg)
		m.options, optionsCmd = m.options.Update(msg)
		cmds.Add(chatCmd, optionsCmd)
		m.Log("Done with update")
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+h":
			return m.switchModel()
		case "ctrl+c":
			switch m.state {
			case tuiStateOptions:
				return m, tea.Quit
				//return m.switchModel()
			}
		}
	}

	// pass through the message to the current model
	var cmd tea.Cmd
	switch m.state {
	case tuiStateChat:
		m.chat, cmd = m.chat.Update(msg)
		cmds.Add(cmd)
	case tuiStateOptions:
		m.options, cmd = m.options.Update(msg)
		cmds.Add(cmd)
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
		return "??"
	}
}
