package ui

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/markdown"
	"github.com/collinvandyck/gpterm/lib/store"
	"github.com/collinvandyck/gpterm/lib/ui/gptea"
	"github.com/google/go-github/v39/github"
	"github.com/gregjones/httpcache"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
)

const (
	defaultChatlogMaxSize = 100
)

type controlModel struct {
	uiOpts
	prompt     tea.Model
	status     tea.Model
	typewriter tea.Model
	textInput  textInput
	backlog    backlog // message backlog loaded from store
	config     config  // persisted config
	ready      bool    // has the terminal initialized
	inflight   bool    // is there a completion in flight
	width      int
	height     int
}

type textInput struct {
	textinput.Model
	active         bool
	credentialName string
}

type config struct {
	store.Config
	query.ClientConfig
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
	if m.textInput.active {
		res += m.textInput.View()
		res += "\n"
	}
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
		m.config.ClientConfig = msg.ClientConfig
		m.config.set = true
		m.client.Update(
			client.WithModel(m.config.ClientConfig.Model),
			client.WithClientContext(int(m.config.ClientConfig.MessageContext)),
		)
		m.Log("Config loaded", "len", len(msg.Config), "client", msg.ClientConfig, "err", msg.Err)

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

	case gptea.EditorRequestMsg:
		if m.ready && !m.inflight {
			cmds.Add(m.spawnEditor(msg.Prompt))
		}

	case gptea.EditorResultMsg:
		if msg.Err != nil {
			cmds.Add(m.error(msg.Err))
			break
		}

	case gptea.ErrorMsg:
		cmds.Add(m.error(msg.Err))

	case gptea.SetCredentialReq:
		m.textInput.active = true
		m.textInput.credentialName = store.CredentialGithubToken
		m.textInput.Model = textinput.New()
		m.textInput.Model.Prompt = msg.Prompt
		m.textInput.Model.EchoMode = textinput.EchoPassword
		m.textInput.Model.Focus()

	case gptea.GistResultMsg:
		seq := []tea.Cmd{}
		if msg.Err != nil {
			seq = append(seq, m.error(msg.Err))
		}
		if msg.NoCredentials {
			seq = append(seq, m.error(errors.New("no GitHub credentials configured")))
			seq = append(seq, gptea.MessageCmd(gptea.SetCredentialReq{
				Prompt: "Enter your GitHub token: ",
				Key:    store.CredentialGithubToken,
			}))
		}
		if msg.URL != "" {
			seq = append(seq, tea.Println())
			role := m.styles.Role(openai.ChatMessageRoleAssistant)
			seq = append(seq, tea.Println(role))
			seq = append(seq, tea.Println("Here's a link to the conversation history: "+msg.URL))
		}
		cmds.Add(tea.Sequence(seq...))

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyF1:
			if m.ready && !m.inflight {
				cmds.Add(m.changeConvoHistory(-1))
			}

		case tea.KeyF2:
			if m.ready && !m.inflight {
				cmds.Add(m.changeConvoHistory(+1))
			}

		case tea.KeyF3:
			if m.ready && !m.inflight {
				cmds.Add(m.cycleClientConfig())
			}
		case tea.KeyF12:
			// gist support isn't ready yet
			// cmds.Add(m.gist)

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

	var textInputCmd tea.Cmd
	m.textInput.Model, textInputCmd = m.textInput.Model.Update(msg)
	cmds.Add(textInputCmd)

	return m, tea.Batch(cmds...)
}

func (m controlModel) gist() tea.Msg {
	ctx := context.Background()
	accessToken, err := m.store.GetCredential(ctx, store.CredentialGithubToken)
	if err != nil {
		return gptea.GistResultMsg{Err: err}
	}
	if accessToken == "" {
		return gptea.GistResultMsg{NoCredentials: true}
	}
	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := &http.Client{
		Transport: &oauth2.Transport{
			Base:   httpcache.NewMemoryCacheTransport(),
			Source: tokenSrc,
		},
	}
	c := github.NewClient(client)
	gist, resp, err := c.Gists.Create(ctx, &github.Gist{
		Description: github.String("gpterm backlog"),
		Public:      github.Bool(false),
		Files: map[github.GistFilename]github.GistFile{
			"gpterm-backlog.md": {
				Filename: github.String("gpterm-backlog.md"),
				Content:  github.String(m.renderBacklogGist()),
			},
		},
	})
	if err != nil {
		return gptea.GistResultMsg{Err: err}
	}
	if resp != nil && resp.StatusCode >= 400 {
		return gptea.GistResultMsg{Err: fmt.Errorf("gist request failed with status code %d", resp.StatusCode)}
	}
	if gist == nil || gist.HTMLURL == nil {
		return gptea.GistResultMsg{Err: fmt.Errorf("gist request failed with no error and no gist")}
	}

	res := gptea.GistResultMsg{URL: *gist.HTMLURL}
	m.Log("Got gist", "url", res.URL)
	var ec *exec.Cmd
	switch os := runtime.GOOS; os {
	case "darwin":
		m.Log("Invoking open")
		ec = exec.Command("/usr/bin/open", *gist.HTMLURL)
	case "linux":
		m.Log("Invoking xdg-open")
		ec = exec.Command("xdg-open", *gist.HTMLURL)
	default:
		m.Log("Unknown OS")
	}
	if ec != nil {
		eco, err := ec.CombinedOutput()
		m.Log("Open result", "err", err, "output", string(eco))
		if err != nil {
			res.Err = fmt.Errorf("failed to open gist in browser: %w\n%s", err, eco)
		}
		res.Opened = err == nil
	}
	return res
}

const editorTemplate = `
# The contents of this file will be written to gpterm 
# as if you had typed it into the prompt.
#
# The first lines of the file beginning with '#' will 
# be ignored when sending the result to gpterm.
#
# Content will be rendered as markdown in the gpterm 
# backlog. Paragraphs should be separated by a blank line.


`

func editorIsVim(editor string) bool {
	editor = strings.TrimSpace(editor)
	editor = strings.ToLower(editor)
	switch editor {
	case "vim", "vi", "nvim":
		return true
	default:
		return false
	}
}

func (m controlModel) spawnEditor(prompt string) tea.Cmd {
	const editorEnv = "EDITOR"
	editor := os.Getenv(editorEnv)
	if editor == "" {
		return gptea.ErrorCmd(fmt.Errorf("Environment variable %q must be set", editorEnv))
	}
	f, err := os.CreateTemp("", "gptea")
	if err != nil {
		return gptea.ErrorCmd(err)
	}
	template := strings.TrimLeft(editorTemplate, " \n")
	io.Copy(f, strings.NewReader(template))
	io.Copy(f, strings.NewReader(prompt))

	args := []string{}
	if editorIsVim(editor) {
		args = append(args, "+$", "-c", "startinsert", "-c", "set ft=markdown syntax=markdown")
	}
	args = append(args, f.Name())
	cmd := exec.Command(editor, args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return gptea.ErrorMsg{Err: err}
		}
		err = f.Close()
		if err != nil {
			return gptea.ErrorMsg{Err: err}
		}
		bs, err := os.ReadFile(f.Name())
		if err != nil {
			return gptea.ErrorMsg{Err: err}
		}
		err = os.Remove(f.Name())
		if err != nil {
			return gptea.ErrorMsg{Err: err}
		}
		text := string(bs)
		text = strings.TrimSpace(text)
		buf := new(bytes.Buffer)
		inHeader := true
		s := bufio.NewScanner(strings.NewReader(text))
		for s.Scan() {
			line := s.Text()
			if inHeader && strings.HasPrefix(line, "#") {
				continue
			}
			inHeader = false
			buf.WriteString(line + "\n")
		}
		text = buf.String()
		text = strings.TrimSpace(text)
		return gptea.EditorResultMsg{Text: text}
	})
}

func (m controlModel) cycleClientConfig() tea.Cmd {
	return func() tea.Msg {
		ctx := m.storeContext()
		m.store.CycleClientConfig(ctx)
		config, err := m.store.GetConfig(ctx)
		if err != nil {
			return gptea.ConfigLoadedMsg{Config: config, Err: err}
		}
		cc, err := m.store.GetClientConfig(ctx)
		return gptea.ConfigLoadedMsg{Config: config, ClientConfig: cc, Err: err}
	}
}

func (m controlModel) changeConvoHistory(delta int) tea.Cmd {
	return func() tea.Msg {
		ctx := m.storeContext()
		val := int(m.config.ClientConfig.MessageContext)
		val += delta
		if val < 1 || val > 20 {
			return gptea.ConversationHistoryMsg{Err: fmt.Errorf("invalid value: %d", val)}
		}
		err := m.store.UpdateClientConfig(ctx, int64(val))
		if err != nil {
			return gptea.ConversationHistoryMsg{Val: val, Err: err}
		}
		config, err := m.store.GetConfig(ctx)
		if err != nil {
			return gptea.ConfigLoadedMsg{Config: config, Err: err}
		}
		cc, err := m.store.GetClientConfig(ctx)
		return gptea.ConfigLoadedMsg{Config: config, ClientConfig: cc, Err: err}
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
	if err != nil {
		return gptea.ConfigLoadedMsg{Config: cfg, Err: err}
	}
	clientCfg, err := m.store.GetClientConfig(ctx)
	return gptea.ConfigLoadedMsg{Config: cfg, ClientConfig: clientCfg, Err: err}
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

func (m controlModel) renderBacklogGist() string {
	buf := bytes.Buffer{}
	for _, msg := range m.backlog.messages {
		role := m.styles.Name(msg.Role)
		content := msg.Content
		buf.WriteString(fmt.Sprintf("### %s\n\n", role))
		buf.WriteString(content)
		buf.WriteString("\n\n")
	}
	re := buf.String()
	re = strings.TrimSpace(re)
	return re
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

	if msg.Content == "" {
		return role
	}
	bs, err := markdown.RenderString(msg.Content, width)
	if err != nil {
		panic(err)
	}
	bs = bytes.TrimSpace(bs)
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
				clientHistory := m.config.ClientConfig.MessageContext
				m.Log("Using client history", "val", clientHistory)
				latest, err := m.store.GetLastMessages(ctx, int(clientHistory))
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
