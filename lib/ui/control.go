package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	markdown "github.com/collinvandyck/go-term-markdown"
	"github.com/collinvandyck/gpterm/lib/term"
	"github.com/collinvandyck/gpterm/lib/ui/command"
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
	typewriter     tea.Model
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
		uiOpts: uiOpts.NamedLogger("control"),
		chat: chatModel{
			uiOpts: uiOpts.NamedLogger("chat"),
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
		m.loadHistory(),
		m.prompt.Init(),
		m.status.Init(),
		m.typewriter.Init(),
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
	tv := m.typewriter.View()
	if strings.TrimSpace(tv) != "" {
		res += tv
		res += "\n"
	}
	res += "\n" // we want a line above our prompt

	res += m.prompt.View()
	res += "\n"
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
		m.Log("Reloaded")
		m.ready = true
		if m.historyLoaded && !m.historyPrinted {
			m.Log("Printing history after reloaded")
			cmds.Add(m.printHistory())
		}

	case redrawMsg:
		if m.historyLoaded && !m.historyPrinted {
			m.Log("Printing history after redraw")
			cmds.Add(m.printHistory())
		}

	case tea.WindowSizeMsg:
		m.Log("Window size changed", "width", msg.Width, "height", msg.Height)
		m.ready = false
		m.historyPrinted = false
		m.width = msg.Width
		m.height = msg.Height
		cmds.Add(m.resetCmd())

	case convoMsg:
		m.Log("Convo switch", "err", msg.err)

	case historyPrinted:
		m.Log("History printed")
		m.historyPrinted = true

	case historyLoaded:
		m.history.addAll(msg.messages, msg.err)
		m.historyLoaded = true
		m.Log("Loaded history", "len", len(m.history.entries))
		if !m.historyPrinted {
			cmds.Add(m.printHistory())
		}

	case command.StreamCompletionReq:
		if msg.Dummy {
			entries := m.history.entries
			if len(entries) <= 1 {
				break
			}
			m.inflight = true
			text := entries[len(entries)-1].Message.Content
			cmds.Add(tea.Sequence(
				m.completeStaticStream(text),
			))
		} else {
			m.inflight = true
			m.history.addUserPrompt(msg.Text)
			m.history.addAssistantPlaceholder()
			cmds.Add(tea.Sequence(
				m.printLastHistories(2),
				m.completeStream(msg.Text),
			))
		}
	case command.StreamCompletion:
	case command.StreamCompletionResult:
		m.inflight = false
		m.history.entries[len(m.history.entries)-1].Message.Content = msg.Text
		m.history.entries[len(m.history.entries)-1].err = msg.Err
		m.history.entries[len(m.history.entries)-1].placeholder = false

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

		default:
		}

	}
	m.prompt = cmds.Update(m.prompt, msg)
	m.status = cmds.Update(m.status, msg)
	m.typewriter = cmds.Update(m.typewriter, msg)

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
			m.Log("Switching to next convo")
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
			m.Log("Switching to prev convo")
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
		func() tea.Msg { m.Log("Clearing scrollback"); term.ClearScrollback(); return nil },
		func() tea.Msg { return reloaded{} },
	)
}

func (m controlModel) printLastHistories(count int) tea.Cmd {
	commands := []tea.Cmd{}
	m.Log("Printing last historic entries", "count", count)
	for i := 0; i < count; i++ {
		buf := bytes.Buffer{}
		he := m.history.entries[len(m.history.entries)-count+i]
		re := m.renderEntry(he)
		re = strings.TrimSpace(re)
		buf.WriteString("\n")
		buf.WriteString(re)
		out := buf.String()
		commands = append(commands, tea.Println(out))
	}
	return tea.Sequence(commands...)
}

func (m controlModel) printLastHistory() tea.Cmd {
	m.Log("Printing last historic entry")
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
	m.Log("Printing history", "len", len(m.history.entries))
	defer m.Log("Done printing")
	buf := bytes.Buffer{}
	for i, he := range m.history.entries {
		re := m.renderEntry(he)
		re = strings.TrimSpace(re)
		buf.WriteString(re)
		if i < len(m.history.entries)-1 {
			buf.WriteString("\n\n")
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
	if entry.placeholder {
		return role
	}
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

func (m controlModel) completeStaticStream(text string) tea.Cmd {
	return func() tea.Msg {
		csm := command.NewStreamCompletion()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
			defer cancel()
			err := func() error {
				pattern := regexp.MustCompile(`\S+|\s+`)
				fields := pattern.FindAllString(text, -1)
				for _, field := range fields {
					err := csm.Write(ctx, field)
					if err != nil {
						return err
					}
					time.Sleep(time.Duration(rand.Int()%25) * time.Millisecond)
				}
				return nil
			}()
			csm.Close(err)
		}()
		return csm
	}
}

func (m controlModel) completeStream(msg string) tea.Cmd {
	return func() tea.Msg {
		csm := command.NewStreamCompletion()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
			defer cancel()
			buf := new(bytes.Buffer) // we'll use this for saving the response
			var usage openai.Usage
			err := func() error {
				latest, err := m.store.GetLastMessages(ctx, m.clientContext)
				if err != nil {
					return fmt.Errorf("load context: %w", err)
				}
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
		m.Log("Completion done", "dur", dur, "ctx", m.clientContext)
		return completion{res.Response.Choices, err}
	}
}

func (m controlModel) loadHistory() tea.Cmd {
	return func() tea.Msg {
		m.Log("Loading history")
		ctx, cancel := context.WithTimeout(context.Background(), m.clientTimeout)
		defer cancel()
		msgs, err := m.store.GetLastMessages(ctx, defaultChatlogMaxSize)
		return historyLoaded{msgs, err}
	}
}
