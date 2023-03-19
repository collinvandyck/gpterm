package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sashabaranov/go-openai"
)

type chatModel struct {
	uiOpts
	viewport        viewport.Model
	textarea        textarea.Model
	spinner         spinner.Model
	styles          styles
	entries         []chatEntry
	err             error
	readyTerm       bool          // when we are ready to render
	readyHist       bool          // true when history is loaded
	readyClient     bool          // true when we are waiting for a client response
	ctxCount        int           // how many messages to send for context
	completeTimeout time.Duration // how long to wait for complete request
}

type chatEntry struct {
	msg query.Message
	err error
}

func newChatModel(uiOpts uiOpts) chatModel {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	res := chatModel{
		uiOpts:          uiOpts,
		readyClient:     true,
		spinner:         spin,
		styles:          newStaticStyles(),
		ctxCount:        5,
		completeTimeout: 30 * time.Second,
	}
	return res
}

func (m chatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.loadHistory(),
		m.spinner.Tick,
	)
}

// complete takes the user input and executes a completion request
// against the client. A completion value will be sent back to the
// ui routine.
func (m chatModel) complete(msg string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		ctx, cancel := m.clientContext()
		defer cancel()
		latest, err := m.store.GetLastMessages(ctx, m.ctxCount)
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
		m.Info("Completion request completed in %s (%d context)", dur, m.ctxCount)
		return completion{res.Response.Choices, err}
	}
}

func (m chatModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := m.clientContext()
		defer cancel()
		msgs, err := m.store.GetLastMessages(ctx, 50)
		return messageHistory{msgs, err}
	}
}

func (m chatModel) clientContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, m.completeTimeout)
}

type completion struct {
	choices []openai.ChatCompletionChoice
	err     error
}

type messageHistory struct {
	messages []query.Message
	err      error
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd   tea.Cmd
		vpCmd   tea.Cmd
		sendCmd tea.Cmd
		spinCmd tea.Cmd
	)
	m.textarea, taCmd = m.textarea.Update(msg)
	m.spinner, spinCmd = m.spinner.Update(msg)

	switch msg := msg.(type) {

	// we have our dimensions. set up the viewport etc.
	case tea.WindowSizeMsg:
		m.Info("Window size changed (w=%d h=%d)", msg.Width, msg.Height)
		var (
			taWidth  = msg.Width
			taHeight = 3
			vpWidth  = msg.Width
			vpHeight = msg.Height - taHeight - 1
		)
		if !m.readyTerm {
			m.textarea = textarea.New()
			m.textarea.Placeholder = "Send a message..."
			m.textarea.Focus()
			m.textarea.Prompt = "┃ "
			m.textarea.CharLimit = 280
			m.textarea.FocusedStyle.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
			m.textarea.ShowLineNumbers = false
			m.viewport = viewport.New(vpWidth, vpHeight)
			m.viewport.SetContent("No conversations yet")
			m.textarea.KeyMap.InsertNewline.SetEnabled(false)
			m.readyTerm = true
			m.Info("Term is ready")
		}
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
		m.textarea.SetWidth(taWidth)
		m.textarea.SetHeight(taHeight)
		m.renderViewport()

	case messageHistory:
		m.Info("Loaded %d historic messages", len(msg.messages))
		for _, qm := range msg.messages {
			m.entries = append(m.entries, chatEntry{msg: qm})
		}
		if msg.err != nil {
			m.entries = append(m.entries, chatEntry{err: msg.err})
		}
		m.readyHist = true
		m.renderViewport()

	case completion:
		for _, choice := range msg.choices {
			m.entries = append(m.entries, chatEntry{
				msg: query.Message{
					Role:    choice.Message.Role,
					Content: choice.Message.Content,
				},
			})
		}
		if msg.err != nil {
			m.entries = append(m.entries, chatEntry{err: msg.err})
		}
		m.readyHist = true
		m.renderViewport()
		m.readyClient = true

	case tea.KeyMsg:
		switch msg.Type {

		// quit
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			return m, tea.Quit

		// todo: message history
		case tea.KeyUp, tea.KeyDown:

		case tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, vpCmd = m.viewport.Update(msg)

		// send a message
		case tea.KeyEnter:
			if !m.readyClient {
				// there's a client request in flight. wait for it
				break
			}
			text := strings.TrimSpace(m.textarea.Value())
			if text != "" {
				m.entries = append(m.entries, chatEntry{
					msg: query.Message{
						Role:    "user",
						Content: m.textarea.Value(),
					},
				})
				sendCmd = m.complete(m.textarea.Value())
				m.renderViewport()
			}
			m.textarea.Reset()
			m.readyClient = false
		}
	case tea.MouseMsg:
		m.viewport, vpCmd = m.viewport.Update(msg)
	case spinner.TickMsg:
	case cursor.BlinkMsg:
	default:
		mtyp := fmt.Sprintf("%T", msg)
		switch mtyp {
		case "cursor.blinkCanceled":
		default:
			// m.Info("%T %v", msg, msg)
		}
	}
	return m, tea.Batch(taCmd, vpCmd, sendCmd, spinCmd)
}

type lineBuilder struct {
	width  int
	buffer bytes.Buffer
}

func (l *lineBuilder) String() string {
	return l.buffer.String()
}

func (l *lineBuilder) Write(line string, md bool) {
	const leftpad = 0
	if md {
		bs := markdown.Render(line, l.width, leftpad)
		line = string(bs)
	} else {
		line = wordwrap.String(line, l.width)
	}
	line = strings.TrimSpace(line)
	l.buffer.WriteString(line)
	l.buffer.WriteString("\n")
}

// When the underlying messages change, this method will render
// those messages into the viewport.
func (m *chatModel) renderViewport() {
	m.Info("Render viewport (%d entries)", len(m.entries))
	b := lineBuilder{width: m.viewport.Width}
	for i, entry := range m.entries {
		b.Write(m.styles.Role(entry.msg.Role), false)
		if entry.err != nil {
			b.Write("error: "+entry.err.Error(), false)
		} else {
			b.Write(entry.msg.Content, true)
		}
		if i < len(m.entries)-1 {
			b.Write("", false)
		}
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m chatModel) View() string {
	if !m.readyTerm {
		return ""
	}
	var res string
	res += m.viewport.View()
	res += "\n"
	if !m.readyClient {
		res += m.spinner.View()
	} else {
		res += m.textarea.View()
	}
	res += "\n"
	return res
}
