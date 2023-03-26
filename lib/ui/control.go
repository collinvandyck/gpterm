package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/collinvandyck/gpterm/lib/term"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultChatlogMaxSize = 100
)

type controlModel struct {
	uiOpts
	prompt         tea.Model
	chat           tea.Model
	status         tea.Model
	history        history
	historyPrinted bool // has the history been printed
	historyLoaded  bool // has the history been loaded
	ready          bool // has the terminal initialized
	help           bool // if we are in help mode
	width          int
	height         int
	inflight       bool
}

type completionReq struct {
	text string
}

type completion struct {
	choices []openai.ChatCompletionChoice
	err     error
}

type historyPrinted struct{}

type reloaded struct {
}

type helpMsg struct {
	help bool
}

func newControlModel(uiOpts uiOpts) controlModel {
	res := controlModel{
		uiOpts: uiOpts.WithLogPrefix("control"),
		chat: chatModel{
			uiOpts: uiOpts.WithLogPrefix("chat"),
			history: history{
				uiOpts:     uiOpts,
				maxEntries: defaultChatlogMaxSize,
			},
			heightOffset: -4,
			styles:       newStaticStyles(),
		},
		history: history{
			maxEntries: defaultChatlogMaxSize,
		},
		prompt: promptModel{
			uiOpts: uiOpts.WithLogPrefix("prompt"),
			height: 3,
		},
		status: newStatusModel(uiOpts.WithLogPrefix("status")),
	}
	return res
}

func (m controlModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadHistory(),
		m.prompt.Init(),
		m.status.Init(),
	)
}

func (m controlModel) View() string {
	if !m.ready {
		return ""
	}
	if m.help {
		return ""
	}
	var res string
	res += m.prompt.View()
	res += m.status.View()
	return res
}

func (m controlModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds commands

	switch msg := msg.(type) {

	case resetHistoryCmd:
		m.history.clear()
		m.historyLoaded = false
		m.historyPrinted = false

	case reloaded:
		m.Info("Reloaded")
		m.ready = true
		if m.historyLoaded {
			cmds.Add(m.printHistory())
		}

	case redrawMsg:
		if m.historyLoaded {
			cmds.Add(m.printHistory())
		}

	case tea.WindowSizeMsg:
		m.Info("Window size changed (w=%d h=%d)", msg.Width, msg.Height)
		m.ready = false
		m.historyPrinted = false
		m.width = msg.Width
		m.height = msg.Height
		cmds.Add(m.resetCmd())

	case convoMsg:
		m.Info("Convo switch err=%v", msg.err)

	case historyPrinted:
		m.Info("History printed")
		m.historyPrinted = true

	case historyLoaded:
		m.history.addAll(msg.messages, msg.err)
		m.historyLoaded = true
		m.Info("Loaded %d history entries", len(m.history.entries))
		if !m.historyPrinted {
			cmds.Add(m.printHistory())
		}

	case completionReq:
		m.inflight = true
		m.history.addUserPrompt(msg.text)
		cmds.Add(tea.Sequence(
			m.printLastHistory(),
			m.complete(msg.text),
		))

	case completion:
		m.inflight = false
		m.history.addCompletion(msg.choices, msg.err)
		cmds.Add(m.printLastHistory())

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlP:
			if !m.inflight {
				cmds.Add(m.prevConvo())
			}

		case tea.KeyCtrlN:
			if !m.inflight {
				cmds.Add(m.nextConvo())
			}

		case tea.KeyCtrlG:
			if true {
				// don't enable this for now
				break
			}
			m.help = !m.help
			if m.help {
				// enter help mode
				return m, tea.Sequence(
					m.helpStatusCmd(),
					tea.EnterAltScreen,
				)
			} else {
				// exit help mode
				return m, tea.Sequence(
					m.helpStatusCmd(),
					tea.ExitAltScreen,
					m.resetCmd(),
					redraw,
				)
			}

		default:
		}

	}
	m.prompt = cmds.Update(m.prompt, msg)
	m.status = cmds.Update(m.status, msg)

	return m, tea.Batch(cmds...)
}

type redrawMsg struct{}

func redraw() tea.Msg {
	return redrawMsg{}
}

type convoMsg struct {
	err error
}

func (m controlModel) nextConvo() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			m.Info("Switching to next convo")
			ctx := context.Background()
			err := m.store.NextConversation(ctx)
			return convoMsg{err: err}
		},
		resetHistory,
		m.loadHistory(),
		m.resetCmd(),
	)
}

func (m controlModel) prevConvo() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			m.Info("Switching to prev convo")
			ctx := context.Background()
			err := m.store.PreviousConversation(ctx)
			return convoMsg{err: err}
		},
		resetHistory,
		m.loadHistory(),
		m.resetCmd(),
	)
}

func (m controlModel) helpStatusCmd() tea.Cmd {
	return func() tea.Msg {
		return helpMsg{
			help: m.help,
		}
	}
}

type resetHistoryCmd struct{}

func resetHistory() tea.Msg {
	return resetHistoryCmd{}
}

func (m controlModel) resetCmd() tea.Cmd {
	return tea.Sequence(
		tea.ClearScreen,
		func() tea.Msg { m.Info("Clearing scrollback"); term.ClearScrollback(); return nil },
		func() tea.Msg { return reloaded{} },
	)
}

func (m controlModel) printLastHistory() tea.Cmd {
	m.Info("Printing last historic entry")
	buf := bytes.Buffer{}
	he := m.history.entries[len(m.history.entries)-1]
	re := m.renderEntry(he)
	buf.WriteString(re)
	out := buf.String()
	return tea.Sequence(
		tea.Println(out),
	)
}

func (m controlModel) printHistory() tea.Cmd {
	m.Info("Printing %d historic entries", len(m.history.entries))
	buf := bytes.Buffer{}
	for i, he := range m.history.entries {
		re := m.renderEntry(he)
		buf.WriteString(re)
		if i < len(m.history.entries)-1 {
			buf.WriteString("\n")
		}
	}
	out := buf.String()
	lines := strings.Split(out, "\n")
	if m.height > len(lines) {
		out = strings.Repeat("\n", m.height-len(lines)) + out
	}
	return tea.Sequence(
		tea.Println(out),
		func() tea.Msg { return historyPrinted{} },
	)
}

func (m controlModel) renderEntry(entry entry) string {
	role := entry.Role
	if entry.err != nil {
		role = "error"
	}
	role = m.styles.Role(role)
	content := m.renderContent(entry)
	return strings.Join([]string{role, content}, "\n")
}

func (m controlModel) renderContent(entry entry) string {
	line := entry.Message.Content
	if entry.err != nil {
		line = fmt.Sprintf("*%v*", entry.err.Error())
	}
	bs := markdown.Render(line, m.width, 0)
	line = string(bs)
	return line
}

// complete takes the user input and executes a completion request
// against the client. A completion value will be sent back to the
// ui routine.
func (m controlModel) complete(msg string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
		defer cancel()
		latest, err := m.store.GetLastMessages(ctx, m.clientContext)
		if err != nil {
			return completion{err: fmt.Errorf("load context: %w", err)}
		}
		res, err := m.client.Complete(ctx, latest, msg)
		if err != nil {
			return completion{err: fmt.Errorf("failed to complete: %w", err)}
		}
		err = m.store.SaveRequestResponse(ctx, res.Req, res.Response)
		if err != nil {
			return completion{err: fmt.Errorf("store resp: %w", err)}
		}
		dur := time.Since(start).Truncate(time.Millisecond)
		m.Info("Completion request completed in %s (%d context)", dur, m.clientContext)
		return completion{res.Response.Choices, err}
	}
}

func (m controlModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		m.Info("Loading history")
		ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
		defer cancel()
		msgs, err := m.store.GetLastMessages(ctx, defaultChatlogMaxSize)
		return historyLoaded{msgs, err}
	}
}
