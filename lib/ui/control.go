package ui

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/collinvandyck/gpterm/lib/ui/gptea"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultChatlogMaxSize = 100
)

type controlModel struct {
	uiOpts
	prompt     tea.Model
	status     tea.Model
	typewriter tea.Model
	backlog    backlog // message backlog loaded from store
	config     config  // persisted config
	ready      bool    // has the terminal initialized
	inflight   bool    // is there a completion in flight
	width      int
	height     int
}

type config struct {
	store.Config
	set bool
}

type backlog struct {
	set      bool
	messages []query.Message
	printed  bool
}

func newControlModel(uiOpts uiOpts) controlModel {
	res := controlModel{
		uiOpts: uiOpts.NamedLogger("control"),
		prompt: promptModel{
			uiOpts: uiOpts.NamedLogger("prompt"),
			height: 3,
		},
		typewriter: typewriterModel{
			uiOpts: uiOpts.NamedLogger("typewriter"),
		},
		status: newStatusModel(uiOpts.NamedLogger("status")),
	}
	return res
}

func (m controlModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadBacklog,
		m.loadConfig,
		m.prompt.Init(),
		m.status.Init(),
		m.typewriter.Init(),
	)
}

func (m controlModel) View() string {
	if !m.ready {
		return ""
	}
	var res string
	res += m.typewriter.View()
	res += "\n"
	res += m.prompt.View()
	res += "\n"
	res += m.status.View()
	return res
}

func (m controlModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds commands

	switch msg := msg.(type) {

	case gptea.WindowSizeMsg:
		m.ready = msg.Ready
		m.width, m.height = msg.Width, msg.Height

	case tea.WindowSizeMsg:
		m.Log("Window size changed", "width", msg.Width, "height", msg.Height)
		m.ready = false
		m.backlog.printed = false
		m.width, m.height = msg.Width, msg.Height
		// when the window changes size we must send this seq
		//
		// 1. send a WindowSizeMsg with ready=false to disable rendering
		// 2. clear the screen and backlog
		// 3. render the backlog
		// 4. send a WindowSizeMsg with ready=true to re-enable rendering
		seq := []tea.Cmd{}
		seq = append(seq, gptea.WindowResized(msg, false))
		seq = append(seq, gptea.ClearScrollback)
		seq = append(seq, m.printBacklog())
		seq = append(seq, gptea.WindowResized(msg, true))
		return m, tea.Sequence(seq...)

	case gptea.ConfigLoadedMsg:
		m.config.Config = msg.Config
		m.config.set = true
		m.Log("Config loaded", "len", len(msg.Config), "err", msg.Err)

	case gptea.BacklogMsg:
		m.Log("Backlog loaded", "len", len(msg.Messages), "err", msg.Err)
		if msg.Err != nil {
			cmds.Add(m.error(msg.Err))
			break
		}
		m.backlog.messages = msg.Messages
		m.backlog.set = true
		return m, m.printBacklog()

	case gptea.BacklogPrintedMsg:
		m.Log("Backlog printed")
		m.backlog.printed = true

	case gptea.ConversationSwitchedMsg:
		m.Log("ConversationSwitchedMsg", "err", msg.Err)
		switch {
		case errors.Is(msg.Err, store.ErrNoMoreConversations):
		case msg.Err != nil:
			cmds.Add(m.error(msg.Err))
		default:
			m.backlog.messages = msg.Messages
			m.backlog.set = true
			m.backlog.printed = false
			seq := []tea.Cmd{}
			seq = append(seq, gptea.ClearScrollback)
			seq = append(seq, m.printBacklog())
			cmds.Add(tea.Sequence(seq...))
		}

	case gptea.StreamCompletionReq:
		m.inflight = true
		um := query.Message{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Text,
		}
		m.backlog.messages = append(m.backlog.messages, um)
		am := query.Message{
			Role: openai.ChatMessageRoleAssistant,
		}
		m.backlog.messages = append(m.backlog.messages, am)
		cmds.Add(tea.Sequence(
			tea.Println(""),
			tea.Println(m.renderMessage(um)),
			tea.Println(m.renderMessage(am)),
			m.completeStream(msg.Text),
		))

	case gptea.StreamCompletionResult:
		m.inflight = false
		if msg.Err != nil {
			cmds.Add(m.error(msg.Err))
			break
		}
		l := len(m.backlog.messages)
		m.backlog.messages[l-1].Content = msg.Text
		extra := len(m.backlog.messages) - defaultChatlogMaxSize
		if extra > 0 {
			m.backlog.messages = m.backlog.messages[extra:]
		}

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlY:
			m.Log("Entering paste mode")

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlU:
			if m.ready && !m.inflight {
				cmds.Add(m.changeConvoHistory(+1))
			}

		case tea.KeyCtrlD:
			if m.ready && !m.inflight {
				cmds.Add(m.changeConvoHistory(-1))
			}

		case tea.KeyCtrlP:
			if m.ready && !m.inflight {
				cmds.Add(m.previous)
			}

		case tea.KeyCtrlN:
			if m.ready && !m.inflight {
				cmds.Add(m.next)
			}

		default:
		}
	}

	m.prompt = cmds.Update(m.prompt, msg)
	m.status = cmds.Update(m.status, msg)
	m.typewriter = cmds.Update(m.typewriter, msg)
	return m, tea.Batch(cmds...)
}

func (m controlModel) changeConvoHistory(delta int) tea.Cmd {
	return func() tea.Msg {
		ctx := m.storeContext()
		val, err := m.store.GetConfigInt(ctx, "chat.message-context", 5)
		if err != nil {
			return gptea.ConversationHistoryMsg{Err: err}
		}
		val += delta
		err = m.store.SetConfigInt(ctx, "chat.message-context", val)
		if err != nil {
			return gptea.ConversationHistoryMsg{Val: val, Err: err}
		}
		config, err := m.store.GetConfig(ctx)
		return gptea.ConfigLoadedMsg{Config: config, Err: err}
	}
}

func (m controlModel) error(err error) tea.Cmd {
	errStr := m.renderMessage(query.Message{
		Role:    "error",
		Content: err.Error(),
	})
	errStr = strings.TrimSpace(errStr)
	return tea.Println("\n" + errStr)
}

func (m controlModel) next() tea.Msg {
	ctx := m.storeContext()
	err := m.store.NextConversation(ctx)
	if err != nil {
		return gptea.ConversationSwitchedMsg{Err: err}
	}
	msgs, err := m.store.GetLastMessages(ctx, defaultChatlogMaxSize)
	return gptea.ConversationSwitchedMsg{Messages: msgs, Err: err}
}

func (m controlModel) previous() tea.Msg {
	ctx := m.storeContext()
	err := m.store.PreviousConversation(ctx)
	if err != nil {
		return gptea.ConversationSwitchedMsg{Err: err}
	}
	msgs, err := m.store.GetLastMessages(ctx, defaultChatlogMaxSize)
	return gptea.ConversationSwitchedMsg{Messages: msgs, Err: err}
}

func (m controlModel) loadConfig() tea.Msg {
	ctx := m.storeContext()
	cfg, err := m.store.GetConfig(ctx)
	return gptea.ConfigLoadedMsg{Config: cfg, Err: err}
}

func (m controlModel) loadBacklog() tea.Msg {
	ctx := m.storeContext()
	msgs, err := m.store.GetLastMessages(ctx, defaultChatlogMaxSize)
	return gptea.BacklogMsg{Messages: msgs, Err: err}
}

func (m controlModel) printBacklog() tea.Cmd {
	if !m.backlog.set || m.backlog.printed {
		m.Log("Not printing backlog", "set", m.backlog.set, "printed", m.backlog.printed)
		return nil
	}
	re := m.renderBacklog()
	reLines := strings.Count(re, "\n") + 1
	extra := m.height - reLines - 5
	if extra > 0 {
		re = strings.Repeat("\n", extra) + re
	}
	return tea.Sequence(
		tea.Println(re),
		gptea.MessageCmd(gptea.BacklogPrintedMsg{}),
	)
}

func (m controlModel) renderBacklog() string {
	start := time.Now()
	buf := bytes.Buffer{}
	for _, msg := range m.backlog.messages {
		re := m.renderMessage(msg)
		re = strings.TrimSpace(re)
		buf.WriteString(re)
		buf.WriteString("\n\n")
	}
	re := buf.String()
	re = strings.TrimSpace(re)
	m.Log("Backlog rendered", "dur", time.Since(start))
	return re
}

func (m controlModel) renderMessage(msg query.Message) string {
	width := m.width
	if width > m.rhsPadding {
		width -= m.rhsPadding
	}
	role := msg.Role
	role = m.styles.Role(role)
	bs := markdown.Render(msg.Content, width, 0)
	sc := bufio.NewScanner(bytes.NewReader(bs))
	rendered := new(bytes.Buffer)
	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimRight(line, " ")
		rendered.WriteString(line + "\n")
	}
	return strings.Join([]string{role, rendered.String()}, "\n")
}

func (m controlModel) completeStream(msg string) tea.Cmd {
	return func() tea.Msg {
		csm := gptea.NewStreamCompletion()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
			defer cancel()
			buf := new(bytes.Buffer) // we'll use this for saving the response
			var usage openai.Usage
			err := func() error {
				clientHistory := m.config.GetChatMessageContext(5)
				m.Log("Using client history", "val", clientHistory)
				latest, err := m.store.GetLastMessages(ctx, clientHistory)
				if err != nil {
					return fmt.Errorf("load context: %w", err)
				}
				m.Log("Using client history context", "len", len(latest))
				streamResult, err := m.client.Stream(ctx, latest, msg)
				if err != nil {
					return fmt.Errorf("failed to complete: %w", err)
				}
				req := streamResult.Req
				err = m.store.SaveRequest(ctx, req)
				if err != nil {
					return err
				}
				res := streamResult.Response
				for {
					sr, err := res.Recv()
					switch {
					case errors.Is(err, io.EOF):
						m.Log("EOF")
						return nil
					case err != nil:
						m.Log("Stream result failure", "err", err)
						return fmt.Errorf("recv: %w", err)
					}

					// each time we get a response we accumulate the usage
					usage.TotalTokens += sr.Usage.TotalTokens
					usage.PromptTokens += sr.Usage.PromptTokens
					usage.CompletionTokens += sr.Usage.CompletionTokens

					content := sr.Choices[0].Delta.Content
					buf.WriteString(content)
					err = csm.Write(ctx, content)
					if err != nil {
						return err
					}
				}
			}()
			buffered := buf.String()
			if err == nil {
				err = m.store.SaveStreamResults(ctx, buffered, usage, err)
			}
			csm.Close(err)
		}()
		return csm
	}
}

func (m controlModel) storeContext() context.Context {
	return context.Background()
}
