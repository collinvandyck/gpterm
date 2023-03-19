package ui

import (
	"time"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

type chatlog struct {
	entries    []entry
	maxEntries int
}

type entry struct {
	query.Message       // an actul message
	err           error // if an error happens we display it in the log
}

func newChatlog(maxEntries int) chatlog {
	return chatlog{
		maxEntries: maxEntries,
	}
}

func (c *chatlog) addUserPrompt(prompt string) {
	c.addMessage(query.Message{
		Timestamp: time.Now(),
		Role:      "user",
		Content:   prompt,
	})
}

func (c *chatlog) addCompletion(choices []openai.ChatCompletionChoice, err error) {
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
	c.entries = append(c.entries, entry)
	if len(c.entries) > c.maxEntries {
		c.entries = c.entries[1:]
	}
}
