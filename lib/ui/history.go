package ui

import (
	"time"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

type history struct {
	uiOpts
	entries    []entry
	maxEntries int
}

type entry struct {
	query.Message // an actul message
	rendered      string
	err           error // if an error happens we display it in the log
	placeholder   bool  // true if we are waiting for a response
}

func (c *history) clear() {
	c.entries = nil
}

func (c *history) addUserPrompt(prompt string) {
	msg := query.Message{
		Timestamp: time.Now(),
		Role:      "user",
		Content:   prompt,
	}
	entry := entry{Message: msg}
	c.addEntry(entry)
}

func (c *history) addAssistantPlaceholder() {
	msg := query.Message{
		Timestamp: time.Now(),
		Role:      "assistant",
		Content:   "",
	}
	entry := entry{Message: msg, placeholder: true}
	c.addEntry(entry)
}

func (c *history) addCompletion(choices []openai.ChatCompletionChoice, err error) {
	c.addCompletionChoices(choices)
	c.addError(err)
}

func (c *history) addCompletionChoices(choices []openai.ChatCompletionChoice) {
	for _, choice := range choices {
		c.addCompletionChoice(choice)
	}
}

func (c *history) addCompletionChoice(choice openai.ChatCompletionChoice) {
	msg := query.Message{
		Timestamp: time.Now(),
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
	}
	entry := entry{Message: msg}
	c.addEntry(entry)
}

func (c *history) addAll(messages []query.Message, err error) {
	c.addMessages(messages)
	c.addError(err)
}

func (c *history) addMessages(msgs []query.Message) {
	for _, m := range msgs {
		c.addMessage(m)
	}
}

func (c *history) addMessage(msg query.Message) {
	entry := entry{Message: msg}
	c.addEntry(entry)
}

func (c *history) addError(err error) {
	if err == nil {
		return
	}
	entry := entry{err: err}
	c.addEntry(entry)
}

func (c *history) addEntry(entry entry) {
	c.entries = append(c.entries, entry)
	if len(c.entries) > c.maxEntries {
		c.entries = c.entries[1:]
	}
}
