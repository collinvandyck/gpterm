package ui

import (
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
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultChatlogMaxSize = 100
)

type chatModel struct {
	uiOpts
	viewport        viewport.Model
	textarea        textarea.Model
	spinner         tea.Model
	render          render
	styles          styles
	chatlog         chatlog
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
	res := chatModel{
		uiOpts:          uiOpts,
		chatlog:         newChatlog(defaultChatlogMaxSize, uiOpts),
		styles:          newStaticStyles(),
		render:          newRender(newStaticStyles()),
		spinner:         newSpinner(uiOpts),
		readyClient:     true,
		ctxCount:        5,
		completeTimeout: 30 * time.Second,
	}
	return res
}

func (m chatModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.loadHistory(),
		m.spinner.Init(),
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
		msgs, err := m.store.GetLastMessages(ctx, m.chatlog.maxEntries)
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
			m.textarea.Placeholder = "..."
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
		m.render.setWidth(vpWidth)
		m.renderViewport()
		m.viewport.GotoBottom()

	case messageHistory:
		m.Info("Loaded %d historic messages", len(msg.messages))
		m.chatlog.addAll(msg.messages, msg.err)
		m.renderViewport()
		m.viewport.GotoBottom()
		m.readyHist = true

	case completion:
		m.chatlog.addCompletion(msg.choices, msg.err)
		m.renderViewport()
		m.viewport.GotoBottom()
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
			if text == "" {
				break
			}
			m.chatlog.addUserPrompt(text)
			m.renderViewport()
			sendCmd = m.complete(m.textarea.Value())
			m.textarea.Reset()
			m.readyClient = false
			m.viewport.GotoBottom()

		case tea.KeyCtrlV:
			m.Info("paste")
		case tea.KeyCtrlG:
			// we'll start off using ctrl-g for inline menu
			m.Info("ctrl-g")
		}
	case tea.MouseMsg:
		m.viewport, vpCmd = m.viewport.Update(msg)
	case spinner.TickMsg:
		m.renderViewport()
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

// When the underlying messages change, this method will render
// those messages into the viewport.
func (m *chatModel) renderViewport() {
	rendered, ok := m.chatlog.render(m.render, m.spinner)
	if ok {
		m.viewport.SetContent(rendered)
	}
}

func (m chatModel) View() string {
	if !m.readyTerm {
		return ""
	}
	var res string
	res += m.viewport.View()
	res += "\n"
	res += m.textarea.View()
	res += "\n"
	return res
}
