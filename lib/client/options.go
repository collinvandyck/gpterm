package client

type Option func(*client)

func WithChatContext(cc ChatContext) Option {
	return func(c *client) {
		c.chatContext = cc
	}
}

func WithContextCount(count int) Option {
	return func(c *client) {
		c.contextCount = count
	}
}
