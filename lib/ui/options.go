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
	var (
		margin = 2
		border = 1
		height = o.height - margin*2 - border*2
		width  = o.width - margin*2 - border*2
	)
	var style = lipgloss.NewStyle().
		Bold(true).
		Blink(true).
		Reverse(false).
		Underline(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#151f18")).
		Width(width).
		Height(height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Margin(margin)

	var inner = lipgloss.Place(20, 30, lipgloss.Center, lipgloss.Center, "hello world")

	return style.Render(inner)
}
