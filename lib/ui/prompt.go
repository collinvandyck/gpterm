package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/errs"
)

var _ tea.Model = promptModel{}

type promptModel struct {
	uiOpts
	ta       textarea.Model
	ready    bool
	height   int
	idx      int    // 0 means current, positive is index in history
	save     string // the current prompt, saved
	inflight bool   // if a client command is in flight
	help     bool   // if we are in help
	width    int
}

type promptHistory struct {
	prompt string
	found  bool
	idx    int
	err    error
}

func (m promptModel) Init() tea.Cmd {
	return cursor.Blink
	//return nil
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		cmds  commands
	)
	m.ta, taCmd = m.ta.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Info("Prompt disabled")
		m.ready = false
		m.width = msg.Width
	case reloaded:
		m.Info("Reloaded")
		if !m.ready {
			m.ta = textarea.New()
			m.ta.Placeholder = "..."
			m.ta.Focus()
			m.ta.Prompt = "┃ "
			m.ta.CharLimit = 4096
			m.ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
			m.ta.ShowLineNumbers = false
			m.ta.KeyMap.InsertNewline.SetEnabled(false)
		}
		m.ta.SetWidth(m.width)
		m.ta.SetHeight(m.height)
		m.Info("prompt width: %d", m.width)
		m.ready = true
	case promptHistory:
		switch {
		case msg.found:
			m.ta.SetValue(msg.prompt)
		case msg.idx == 0:
			m.ta.SetValue(m.save)
		}
		m.idx = msg.idx
	case completion:
		m.inflight = false

	case helpMsg:
		m.help = msg.help

	case tea.KeyMsg:
		if m.help {
			break
		}
		switch msg.Type {
		case tea.KeyUp:
			if m.idx == 0 {
				m.save = m.ta.Value()
			}
			cmds.Add(m.getPrevious(1))
		case tea.KeyDown:
			switch m.idx {
			case 1:
				m.ta.SetValue(m.save)
				m.idx = 0
			case 0: // do nothing
			default:
				cmds.Add(m.getPrevious(-1))
			}
		case tea.KeyEnter:
			if m.inflight {
				break
			}
			text := strings.TrimSpace(m.ta.Value())
			if text == "" {
				break
			}
			cmds = append(cmds, m.complete(text))
			m.ta.Reset()
			m.inflight = true
		}
	}
	return m, cmds.BatchWith(taCmd)
}

func (m promptModel) View() string {
	if !m.ready {
		return ""
	}
	return m.ta.View() + "\n"
}

func (m promptModel) getPrevious(inc int) tea.Cmd {
	var (
		idx   = m.idx
		taVal = m.ta.Value()
	)
	return func() tea.Msg {
		ctx := context.Background()
		for {
			idx += inc
			msg, err := m.store.GetPreviousMessageForRole(ctx, "user", idx)
			switch {
			case errs.IsDBNotFound(err):
				return promptHistory{
					idx: idx,
				}
			case err != nil:
				return promptHistory{
					err: err,
					idx: idx,
				}
			case msg.Content != taVal:
				return promptHistory{
					prompt: msg.Content,
					found:  true,
					idx:    idx,
				}
			}
		}
	}
}

func (m promptModel) complete(msg string) tea.Cmd {
	return func() tea.Msg {
		return completionReq{msg}
	}
}

func (m promptModel) update(msg tea.Msg) (promptModel, tea.Cmd) {
	res, cmd := m.Update(msg)
	return res.(promptModel), cmd
}

func (m promptModel) SetValue(val string) {
	m.ta.SetValue(val)
}

func (m promptModel) Value() string {
	return m.ta.Value()
}
