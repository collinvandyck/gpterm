package ui

import tea "github.com/charmbracelet/bubbletea"

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
func (o optionsModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return o, nil
}

// View implements tea.Model.
func (o optionsModel) View() string {
	return "options"
}
