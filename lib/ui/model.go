package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// model is our custom model interface that allows models to
// return themselves in the Update method.
//
// This allows us to have a model that can be composed of other models,
// without the awkward type assertions that would be otherwise
// required to set the model after the Update method returns.
type model interface {
	Init() tea.Cmd
	Update(tea.Msg) (model, tea.Cmd)
	View() string
}

// baseModel is a model that allows common data to be tracked.
type baseModel struct {
	width  int
	height int
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(tea.Msg) (model, tea.Cmd) {
	return m, nil
}

func (m baseModel) View() string {
	return ""
}
