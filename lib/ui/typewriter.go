package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/markdown"
	"github.com/collinvandyck/gpterm/lib/ui/gptea"
)

type typewriterModel struct {
	uiOpts
	width       int
	height      int
	data        string   // the raw data
	rendered    []string // the rendered markdown
	maxRendered int      // the max rendered lines so far
}

func (m typewriterModel) Init() tea.Cmd {
	return nil
}

func (m typewriterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds commands

	switch msg := msg.(type) {

	case gptea.StreamCompletionReq:
		m.reset()

	case gptea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case gptea.StreamCompletion:
		words := msg.Next()
		for _, str := range words {
			m.write(str)
		}
		m.render()
		if !msg.Done() {
			tick := tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg { return msg })
			cmds.Add(tick)
			break
		}
		// print all rendered lines to the console before we stop being rendered
		for _, r := range m.rendered {
			cmds.Add(tea.Println(r))
		}

		// signal to control that we are done with the stream and these are the results
		scr := gptea.StreamCompletionResult{
			Err:  msg.Err(),
			Text: m.data,
		}
		cmds.Add(func() tea.Msg { return scr })
		m.reset()

	}
	return m, cmds.Sequence()

}

func (m typewriterModel) View() string {
	style := lipgloss.NewStyle() //.Background(lipgloss.Color("#333333"))
	if len(m.rendered) == 0 {
		return ""
	}
	lines := m.rendered
	for i := len(lines); i < m.maxRendered; i++ {
		lines = append(lines, "")
	}
	lines = m.truncLines(lines)
	res := strings.Join(lines, "\n")
	res = style.Render(res)
	if strings.TrimSpace(res) != "" {
		res += "\n"
	}
	return res
}

func (m *typewriterModel) write(part string) {
	m.data += part
}

func (m *typewriterModel) render() {
	data := m.data
	stanzas := m.countCodeStanzas()
	if stanzas%2 == 1 {
		data += "\n" + codeStanza
	}
	width := m.width
	if width > m.rhsPadding {
		width -= m.rhsPadding
	}
	bs, _ := markdown.RenderString(data, width)
	re := string(bs)
	re = strings.TrimSpace(re)
	m.rendered = strings.Split(re, "\n")
	if len(m.rendered) > m.maxRendered {
		m.maxRendered = len(m.rendered)
	}
}

const codeStanza = "```"

func (m *typewriterModel) countCodeStanzas() (count int) {
	data := m.data
	for {
		idx := strings.Index(data, codeStanza)
		if idx == -1 {
			break
		}
		data = data[idx+len(codeStanza):]
		count++
	}
	return count
}

func (m *typewriterModel) reset() {
	m.data = ""
	m.rendered = nil
	m.maxRendered = 0
}

func (m typewriterModel) truncLines(lines []string) []string {
	max := m.height - 6 // 3 for prompt, one for status bar, and one for padding
	if max < 1 {
		max = 1
	}
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return lines
}
