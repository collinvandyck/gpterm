package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionsModel struct {
	uiOpts
	width  int
	height int
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
	case tea.WindowSizeMsg:
		o.width, o.height = msg.Width, msg.Height
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
		Width(o.width).
		Height(o.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return style.Render("Jeeves reporting for duty")
}
