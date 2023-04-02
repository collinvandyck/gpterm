package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/collinvandyck/gpterm/db/query"
)

var _ tea.Model = chatModel{}

type chatModel struct {
	uiOpts
	viewport     viewport.Model
	styles       styles
	history      history
	ready        bool
	heightOffset int
	width        int
	buf          *bytes.Buffer
}

type historyLoaded struct {
	messages []query.Message
	err      error
}

func (m chatModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadHistory(),
	)
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		cmds  commands
	)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height+m.heightOffset)
			m.viewport.SetContent("No conversations yet")
			m.viewport.KeyMap = viewport.KeyMap{
				PageDown: key.NewBinding(
					key.WithKeys("pgdown"),
				),
				PageUp: key.NewBinding(
					key.WithKeys("pgup"),
				),
			}
			m.ready = true
		}
		m.width = msg.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height + m.heightOffset
		m.forceRender()
		m.viewport.GotoBottom()

	case historyLoaded:
		m.Log("Loaded chatlog", "count", len(msg.messages))
		m.history.addAll(msg.messages, msg.err)
		m.render()
		m.viewport.GotoBottom()

	case completionReq:
		m.history.addUserPrompt(msg.text)
		m.render()
		m.viewport.GotoBottom()

	case completion:
		m.history.addCompletion(msg.choices, msg.err)
		m.render()
		m.viewport.GotoBottom()

	case tea.KeyMsg:
	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseLeft:
			//m.Info("Disabling mouse on click")
			//cmds.Add(tea.DisableMouse)
		}
	}

	return m, cmds.BatchWith(vpCmd)
}

func (m chatModel) View() string {
	if !m.ready {
		return ""
	}
	return m.viewport.View()
}

func (m chatModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
		defer cancel()
		msgs, err := m.store.GetLastMessages(ctx, m.history.maxEntries)
		return historyLoaded{msgs, err}
	}
}

func (m *chatModel) forceRender() {
	m.doRender(true)
}

func (m *chatModel) render() {
	m.doRender(false)
}

func (m *chatModel) doRender(force bool) {
	m.Log("Rendering chat", "force", force)
	if m.buf == nil {
		m.buf = bytes.NewBuffer(make([]byte, 0, 4096))
	}
	m.buf.Reset()
	for i, entry := range m.history.entries {
		rendered := entry.rendered
		if force || rendered == "" {
			rendered = m.renderEntry(entry)
			m.history.entries[i].rendered = rendered
		}
		m.buf.WriteString(rendered)
		if i < len(m.history.entries)-1 {
			m.buf.WriteString("\n")
		}
	}
	m.viewport.SetContent(m.buf.String())
}

func (m chatModel) renderEntry(entry entry) string {
	role := entry.Role
	if entry.err != nil {
		role = "error"
	}
	role = m.styles.Role(role)
	content := m.renderContent(entry)
	return strings.Join([]string{role, content}, "\n")
}

func (m chatModel) renderContent(entry entry) string {
	line := entry.Message.Content
	if entry.err != nil {
		line = fmt.Sprintf("*%v*", entry.err.Error())
	}
	bs := markdown.Render(line, m.width, 0)
	line = string(bs)
	return line
}
