package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/sqlite"
	"github.com/collinvandyck/gpterm/lib/ui/options"
)

type optionsModel struct {
	uiOpts
	width    int
	height   int
	options  []option
	selected int
	active   bool // whether or not the option is active
}

type option struct {
	name  string
	model options.Interface
}

func newOptionsModel(opts uiOpts) optionsModel {
	model := optionsModel{
		uiOpts: opts.NamedLogger("options"),
	}
	model.options = []option{
		{
			name:  "api key",
			model: options.NewAPIKeyModel(),
		},
		{
			name:  "something else",
			model: options.NewAPIKeyModel(),
		},
	}
	return model
}

func (o *optionsModel) move(direction int) {
	o.selected = (len(o.options) + o.selected + direction) % len(o.options)
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
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.width, o.height = msg.Width, msg.Height
	case tea.KeyMsg:
		// navigation in the menu
		if !o.active {
			switch msg.String() {
			case "j", "down":
				o.move(1)
			case "k", "up":
				o.move(-1)
			case "enter":
				o.active = !o.active
				if o.active {
					cmds = append(cmds, o.options[o.selected].model.Init())
				}
			}
		} else {
			// esc is how we back out of an active option
			switch msg.String() {
			case "esc":
				o.active = !o.active
			}
		}
	case optionTick:
		return o, o.tick()
	}
	if o.active {
		// pass through the message to the active option
		var cmd tea.Cmd
		o.options[o.selected].model, cmd = o.options[o.selected].model.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
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
	if o.active {
		var rhsModel = o.options[o.selected].model.View()
		rhs.WriteString(rhsModel)
	} else {
		rhs.WriteString("...")
	}

	lhsRhs := lipgloss.JoinHorizontal(lipgloss.Top, lhs.String(), rhs.String())
	doc.WriteString(lhsRhs)

	docStyle := lipgloss.NewStyle().
		Margin(2).
		MaxWidth(o.width - 2)

	return docStyle.Render(doc.String())
}
