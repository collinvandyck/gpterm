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
	width    int
	height   int
	options  []option
	selected int
}

type option struct {
	name  string
	model optionInterface
}

func newOptionsModel(opts uiOpts) optionsModel {
	model := optionsModel{
		uiOpts: opts.NamedLogger("options"),
	}
	model.options = []option{
		{
			name:  "api key",
			model: newApiKeyOption("enter api key..."),
		},
		{
			name:  "something else",
			model: newApiKeyOption("enter something else..."),
		},
	}
	return model
}

func (o *optionsModel) move(direction int) tea.Cmd {
	o.selected = (len(o.options) + o.selected + direction) % len(o.options)
	return o.options[o.selected].model.Init()
}

// Init implements tea.Model.
func (o optionsModel) Init() tea.Cmd {
	return tea.Batch(
		o.tick(),
		o.options[o.selected].model.Init(),
	)
}

type optionTick struct{}

func (o optionsModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return optionTick{}
	})
}

// Update implements tea.Model.
func (o optionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.width, o.height = msg.Width, msg.Height
	case tea.KeyMsg:
		// navigation in the menu
		switch msg.String() {
		case "ctrl+p":
			return o, o.move(-1)
		case "ctrl+n":
			return o, o.move(-1)
		}
	case optionTick:
		return o, o.tick()
	}
	// pass through the message to the active option
	var cmd tea.Cmd
	o.options[o.selected].model, cmd = o.options[o.selected].model.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if o.quitAfterRender {
		return o, tea.Quit
	}
	if len(cmds) == 0 {
		return o, nil
	}
	return o, tea.Sequence(cmds...)
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
	doc.WriteString("\n")

	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	doc.WriteString(strings.Repeat(divider.Render("─"), o.width))
	doc.WriteString("\n")

	helpStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333")).
		Foreground(lipgloss.Color("#dddddd")).
		Width(o.width)
	help := helpStyle.Render("ctrl+p: up, ctrl+n: down, enter: select, esc: back out")
	doc.WriteString(help)
	doc.WriteString("\n")
	doc.WriteString(strings.Repeat(divider.Render("─"), o.width))
	doc.WriteString("\n")

	doc.WriteString("\n")
	var lhs strings.Builder
	var styleListItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#abc")).
		PaddingRight(5).
		MarginRight(2).
		ColorWhitespace(true)
	var styleListItemSelected = styleListItem.Copy().Foreground(lipgloss.Color("#333")).Background(lipgloss.Color("#abc"))
	for i, item := range o.options {
		if o.selected == i {
			lhs.WriteString(styleListItemSelected.Render(item.name))
		} else {
			lhs.WriteString(styleListItem.Render(item.name))
		}
		lhs.WriteString("\n")
	}

	var rhs strings.Builder
	var rhsModel = o.options[o.selected].model.View()
	rhs.WriteString(rhsModel)

	lhsRhs := lipgloss.JoinHorizontal(lipgloss.Top, lhs.String(), rhs.String())
	doc.WriteString(lhsRhs)

	docStyle := lipgloss.NewStyle().
		Margin(2).
		MaxWidth(o.width - 2)

	return docStyle.Render(doc.String())
}
