package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type optionsModel struct {
	uiOpts
	width   int
	height  int
	options options
}

type option struct {
	name   string
	active bool
}

type options []option

func (o options) move(direction int) {
	for i, option := range o {
		if option.active {
			o[i].active = false
			o[(len(o)+i+direction)%len(o)].active = true
			return
		}
	}
}

func newOptionsModel(opts uiOpts) optionsModel {
	model := optionsModel{
		uiOpts: opts.NamedLogger("options"),
	}
	model.options = options{
		{
			name:   "api key",
			active: true,
		},
		{
			name: "two",
		},
		{
			name: "three",
		},
	}
	return model
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
		case "j", "down":
			o.options.move(1)
		case "k", "up":
			o.options.move(-1)
		}
	}
	return o, nil
}

// View implements tea.Model.
func (o optionsModel) View() string {
	var doc strings.Builder
	var styleHeader = lipgloss.NewStyle().Bold(true)

	header1 := styleHeader.Render("gpterm options menu")
	header2 := styleHeader.Render(time.Now().Truncate(time.Second).String())
	header := lipgloss.JoinVertical(lipgloss.Top, header1, header2)
	doc.WriteString(header)
	doc.WriteString("\n\n")

	var styleListItem = lipgloss.NewStyle().Foreground(lipgloss.Color("#abc"))
	var styleListItemActive = styleListItem.Copy().Foreground(lipgloss.Color("#fff")).Background(lipgloss.Color("#abc"))
	for _, item := range o.options {
		if item.active {
			doc.WriteString(styleListItemActive.Render(item.name))
		} else {
			doc.WriteString(styleListItem.Render(item.name))
		}
		doc.WriteString("\n")
	}

	return lipgloss.NewStyle().Margin(2).Render(doc.String())
}
