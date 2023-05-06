package client

import (
	"github.com/collinvandyck/gpterm/lib/log"
)

type Option func(*client, *roundTripper)

func WithRequestLogger(log log.Logger) Option {
	return func(c *client, rt *roundTripper) {
		rt.log = log
	}
}

func WithModel(model string) Option {
	return func(c *client, rt *roundTripper) {
		c.model = model
	}
}

func WithClientContext(clientContext int) Option {
	return func(c *client, rt *roundTripper) {
		c.clientContext = clientContext
	}
}
