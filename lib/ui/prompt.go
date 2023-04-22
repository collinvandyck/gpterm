package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/errs"
	"github.com/collinvandyck/gpterm/lib/ui/gptea"
)

type promptModel struct {
	uiOpts
	ta           textarea.Model
	initialized  bool
	ready        bool
	height       int
	idx          int    // 0 means current, positive is index in history
	save         string // the current prompt, saved
	inflight     bool   // if a client command is in flight
	width        int
	editorPrompt string // saves the editor value
}

type promptHistory struct {
	prompt string
	found  bool
	idx    int
	err    error
}

func (m promptModel) Init() tea.Cmd {
	return cursor.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		cmds  commands
	)
	m.ta, taCmd = m.ta.Update(msg)

	switch msg := msg.(type) {

	case gptea.WindowSizeMsg:
		m.Log("WindowSizeMsg")
		if !msg.Ready {
			m.ready = false
			break
		}
		if !m.initialized {
			m.Log("Building prompt")
			m.ta = textarea.New()
			m.ta.Placeholder = "..."
			cmds.Add(m.ta.Focus())
			m.ta.Prompt = "┃ "
			m.ta.CharLimit = 4096
			m.ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
			m.ta.ShowLineNumbers = false
			m.ta.KeyMap.InsertNewline.SetEnabled(false)
			m.initialized = true
		}
		m.width = msg.Width
		m.ta.SetWidth(m.width)
		m.ta.SetHeight(m.height)
		m.ready = true

	case promptHistory:
		switch {
		case msg.found:
			m.ta.SetValue(msg.prompt)
			m.idx = msg.idx
		case msg.idx == 0:
			m.ta.SetValue(m.save)
		}

	case gptea.SetCredentialReq:
		m.ta.Blur()

	case gptea.SetCredentialRes:
		m.ta.Focus()

	case gptea.StreamCompletionResult:
		m.inflight = false

	case gptea.ConversationSwitchedMsg:
		m.idx = 0

	case gptea.EditorResultMsg:
		text := strings.TrimSpace(msg.Text)
		taval := strings.TrimSpace(m.ta.Value())
		if text == taval {
			break
		}
		if text != "" {
			req := gptea.MessageCmd(gptea.StreamCompletionReq{Text: text})
			cmds.Add(req)
			m.ta.Reset()
			m.inflight = true
		}

	case tea.KeyMsg:
		if !m.ta.Focused() {
			break
		}
		switch msg.Type {

		case tea.KeyCtrlY:
			if m.ready && !m.inflight {
				cmds.Add(gptea.StartEditorCmd(m.ta.Value()))
			}

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
			if !m.ready || m.inflight {
				break
			}
			text := strings.TrimSpace(m.ta.Value())
			if text != "" {
				req := gptea.MessageCmd(gptea.StreamCompletionReq{Text: text})
				cmds.Add(req)
				m.ta.Reset()
				m.inflight = true
			}
		}
	}
	return m, cmds.BatchWith(taCmd)
}

func (m promptModel) View() string {
	if !m.ready {
		return ""
	}
	return m.ta.View()
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

func (m promptModel) update(msg tea.Msg) (promptModel, tea.Cmd) {
	res, cmd := m.Update(msg)
	return res.(promptModel), cmd
}
