package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/sqlite"
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
			name: "preludes",
		},
		{
			name: "three",
		},
	}
	return model
}

// Init implements tea.Model.
func (o optionsModel) Init() tea.Cmd {
	return o.tick()
}

type optionTick struct{}

func (o optionsModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return optionTick{}
	})
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
	case optionTick:
		return o, o.tick()
	}
	//return o, tea.Quit
	return o, nil
}

// View implements tea.Model.
func (o optionsModel) View() string {
	var doc strings.Builder
	var styleHeader = lipgloss.NewStyle().Bold(true)

	header1 := styleHeader.Render("gpterm options menu")
	header2 := styleHeader.Render(time.Now().Truncate(time.Second).String())
	header3 := styleHeader.Render(fmt.Sprintf("cgo enabled: %v", sqlite.CGO_ENABLED))
	header := lipgloss.JoinVertical(lipgloss.Top, header1, header2, header3)

	doc.WriteString(header)
	doc.WriteString("\n\n")

	var lhs strings.Builder
	var styleListItem = lipgloss.NewStyle().Foreground(lipgloss.Color("#abc")).MarginRight(5)
	var styleListItemActive = styleListItem.Copy().Foreground(lipgloss.Color("#333")).Background(lipgloss.Color("#abc"))
	for _, item := range o.options {
		if item.active {
			lhs.WriteString(styleListItemActive.Render(item.name))
		} else {
			lhs.WriteString(styleListItem.Render(item.name))
		}
		lhs.WriteString("\n")
	}
	var rhs strings.Builder
	for _, item := range o.options {
		if !item.active {
			continue
		}
		rhs.WriteString(item.name)
	}
	lhsRhs := lipgloss.JoinHorizontal(lipgloss.Top, lhs.String(), rhs.String())
	doc.WriteString(lhsRhs)

	return lipgloss.NewStyle().Margin(2).Render(doc.String())
}
