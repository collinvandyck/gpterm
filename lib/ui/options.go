package ui

import (
	"github.com/collinvandyck/gpterm/lib/log"
)

type Option func(*console)

func WithLogger(logger log.Logger) Option {
	return func(c *console) {
		c.Logger = logger
	}
}
