package client

import "github.com/collinvandyck/gpterm/lib/log"

type Option func(*client, *roundTripper)

func WithRequestLogger(log log.Logger) Option {
	return func(c *client, rt *roundTripper) {
		rt.log = log
	}
}
