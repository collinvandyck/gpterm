package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionsModel struct {
	uiOpts
}

func newOptionsModel(opts uiOpts) optionsModel {
	return optionsModel{
		uiOpts: opts.NamedLogger("options"),
	}
}

// Init implements tea.Model.
func (o optionsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (o optionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		}
	}
	return o, nil
}

// View implements tea.Model.
func (o optionsModel) View() string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(20).
		PaddingBottom(20).
		PaddingLeft(16).
		PaddingRight(16).
		Width(22)
	return style.Render("Hello, kitty")
}
