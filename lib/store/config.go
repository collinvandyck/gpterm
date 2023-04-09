package store

import (
	"errors"
	"strconv"

	"github.com/collinvandyck/gpterm/db/query"
)

type Config []query.Config

func (c Config) Int(key string) (int, error) {
	for _, v := range c {
		if v.Name == key {
			return strconv.Atoi(v.Value)
		}
	}
	return 0, errors.New("config not found")
}

func (c Config) GetChatMessageContext(defaultValue int) int {
	val, err := c.Int("chat.message-context")
	if err != nil {
		return defaultValue
	}
	return val
}
