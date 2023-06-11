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
	ready    bool
	data     *optionsData
}

type option struct {
	name  string
	model model
}

// represents the options that can be configured through the options UI
type optionsData struct {
	apiKey string
}

func newOptionsModel(opts uiOpts) optionsModel {
	model := optionsModel{
		uiOpts: opts.NamedLogger("options"),
	}
	model.options = []option{
		{
			name:  "General",
			model: newGeneralOptions(),
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
	return tea.Sequence(
		o.load(),
		o.tick(),
		o.options[o.selected].model.Init(),
	)
}

func (o optionsModel) load() tea.Cmd {
	return func() tea.Msg {
		return &optionsData{
			apiKey: "f",
		}
	}
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
	case *optionsData:
		o.ready = true
		o.data = msg
		// pass through the data to each option
		for i, opt := range o.options {
			var cmd tea.Cmd
			o.options[i].model, cmd = opt.model.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return o, tea.Sequence(cmds...)
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
	if !o.ready {
		return "loading..."
	}
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
	help := helpStyle.Render("Ctrl+p/n: switch sections | Tab/S-Tab: Cycle fields | Ctrl+c: cancel")
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
	doc.WriteString("\n\n")

	doc.WriteString(fmt.Sprintf("%#+v", o.data))

	docStyle := lipgloss.NewStyle().
		Margin(2).
		MaxWidth(o.width - 2)

	return docStyle.Render(doc.String())
}
