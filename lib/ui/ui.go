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
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sashabaranov/go-openai"
)

type UI interface {
	// blocks until the program exits
	Run(ctx context.Context) error
}

func New(store *store.Store, client client.Client, opts ...Option) UI {
	console := &console{
		Logger: log.Discard,
		store:  store,
		client: client,
		styles: newStaticStyles(),
	}
	for _, o := range opts {
		o(console)
	}
	return console
}

type console struct {
	log.Logger
	store  *store.Store
	client client.Client
	styles styles
}

func (t *console) Run(ctx context.Context) error {
	t.Info("Starting")
	defer t.Info("Exiting")
	model, err := t.chatModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

func (t *console) chatModel() (tea.Model, error) {
	return newChatModel(t), nil
}

type chatEntry struct {
	msg query.Message
	err error
}

type chatModel struct {
	log.Logger
	console  *console
	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model

	entries     []chatEntry
	err         error
	readyTerm   bool // when we are ready to render
	readyHist   bool // true when history is loaded
	readyClient bool // true when we are waiting for a client response
}

// https://github.com/charmbracelet/bubbletea/blob/master/examples/pager/main.go
// https://github.com/charmbracelet/bubbles
func newChatModel(console *console) chatModel {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return chatModel{
		Logger:      console.Logger,
		console:     console,
		readyClient: true,
		spinner:     spin,
	}
}

func (m chatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.loadHistory(),
		m.spinner.Tick,
	)
}

func (m chatModel) sendMessage(msg string) tea.Cmd {
	return func() tea.Msg {
		m.Info("Sending message")
		ctx, cancel := m.clientContext()
		defer cancel()
		res, err := m.console.client.Complete(ctx, msg)
		if err != nil {
			m.Error("Failed to complete text: %v", err)
			return messageResponses{err: err}
		}
		m.Info("Got %d message responses", len(res.Response.Choices))
		return messageResponses{res.Response.Choices, err}
	}
}

func (m chatModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		m.Info("Loading history")
		ctx, cancel := m.clientContext()
		defer cancel()
		msgs, err := m.console.store.GetLastMessages(ctx, 50)
		m.Info("Loaded %d messages err=%v", len(msgs), err)
		return messageHistory{msgs, err}
	}
}

func (m chatModel) clientContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 1*time.Minute)
}

type messageResponses struct {
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
		m.Info("Window size changed: %#+v", msg)
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
		}
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
		m.textarea.SetWidth(taWidth)
		m.textarea.SetHeight(taHeight)
		m.updateViewport()

	case messageHistory:
		m.Info("Got message history update %d messages err=%v", len(msg.messages), msg.err)
		for _, qm := range msg.messages {
			m.entries = append(m.entries, chatEntry{msg: qm})
		}
		if msg.err != nil {
			m.entries = append(m.entries, chatEntry{err: msg.err})
		}
		m.readyHist = true
		m.updateViewport()

	case messageResponses:
		m.Info("Got %d message responses err=%v", len(msg.choices), msg.err)
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
		m.updateViewport()
		m.readyClient = true

	case tea.KeyMsg:
		m.Info("Got key: %s", msg.String())
		switch msg.Type {

		// quit
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyUp, tea.KeyDown:

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
				sendCmd = m.sendMessage(m.textarea.Value())
				m.updateViewport()
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
			m.Info("%T %v", msg, msg)
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

// updates the viewport with the model's current entries
func (m *chatModel) updateViewport() {
	m.Info("Updating viewport entries=%d", len(m.entries))
	b := lineBuilder{width: m.viewport.Width}
	for i, entry := range m.entries {
		b.Write(m.console.styles.Role(entry.msg.Role), false)
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
