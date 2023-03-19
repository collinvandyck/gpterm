package ui

import "io"

type Option func(*console)

func WithLogWriter(w io.Writer) Option {
	return func(c *console) {
		c.logWriter = w
	}
}
