package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/collinvandyck/gpterm/lib/log"
	"github.com/collinvandyck/gpterm/lib/store"
)

type UI interface {
	// blocks until the program exits
	Run(ctx context.Context) error
}

func New(store *store.Store, client client.Client, opts ...Option) UI {
	console := &console{
		uiOpts: uiOpts{
			Logger: log.Discard,
			store:  store,
			client: client,
			styles: newStaticStyles(),
		},
	}
	for _, o := range opts {
		o(console)
	}
	return console
}

type uiOpts struct {
	log.Logger
	store  *store.Store
	client client.Client
	styles styles
}

type console struct {
	uiOpts
}

func (t *console) Run(ctx context.Context) error {
	model := newChatModel(t.uiOpts)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
