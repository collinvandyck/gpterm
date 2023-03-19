package ui

import (
	"bytes"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

type chatlog struct {
	uiOpts
	entries    []entry
	maxEntries int
	dirty      bool
}

type entry struct {
	query.Message       // an actul message
	err           error // if an error happens we display it in the log
	rendered      string
	spin          bool // if we should add a spinner next to this entry
}

func newChatlog(maxEntries int, uiOpts uiOpts) chatlog {
	return chatlog{
		uiOpts:     uiOpts,
		maxEntries: maxEntries,
	}
}

func (c *chatlog) render(render render, spinner tea.Model) (string, bool) {
	if !c.dirty {
		return "", false
	}
	c.dirty = false
	buf := new(bytes.Buffer)
	for i, entry := range c.entries {
		if entry.spin {
			c.dirty = true
		}
		if entry.rendered == "" {
			entry.rendered = render.renderEntry(entry, spinner)
		}
		buf.WriteString(entry.rendered)
		if i < len(c.entries)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String(), true
}

func (c *chatlog) removeSpin() {
	for i := range c.entries {
		c.entries[i].spin = false
	}
}

func (c *chatlog) addUserPrompt(prompt string) {
	msg := query.Message{
		Timestamp: time.Now(),
		Role:      "user",
		Content:   prompt,
	}
	entry := entry{Message: msg, spin: true}
	c.addEntry(entry)
}

func (c *chatlog) addCompletion(choices []openai.ChatCompletionChoice, err error) {
	c.removeSpin()
	c.addCompletionChoices(choices)
	c.addError(err)
}

func (c *chatlog) addCompletionChoices(choices []openai.ChatCompletionChoice) {
	for _, choice := range choices {
		c.addCompletionChoice(choice)
	}
}

func (c *chatlog) addCompletionChoice(choice openai.ChatCompletionChoice) {
	msg := query.Message{
		Timestamp: time.Now(),
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
	}
	entry := entry{Message: msg}
	c.addEntry(entry)
}

func (c *chatlog) addAll(messages []query.Message, err error) {
	c.addMessages(messages)
	c.addError(err)
}

func (c *chatlog) addMessages(msgs []query.Message) {
	for _, m := range msgs {
		c.addMessage(m)
	}
}

func (c *chatlog) addMessage(msg query.Message) {
	entry := entry{Message: msg}
	c.addEntry(entry)
}

func (c *chatlog) addError(err error) {
	if err == nil {
		return
	}
	entry := entry{err: err}
	c.addEntry(entry)
}

func (c *chatlog) addEntry(entry entry) {
	c.dirty = true
	c.entries = append(c.entries, entry)
	if len(c.entries) > c.maxEntries {
		c.entries = c.entries[1:]
	}
}
