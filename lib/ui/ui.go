package ui

import (
	"context"
	"time"

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
			Logger:        log.Discard,
			store:         store,
			client:        client,
			styles:        newStaticStyles(),
			clientTimeout: time.Minute,
			clientContext: 5,
		},
	}
	for _, o := range opts {
		o(console)
	}
	return console
}

type uiOpts struct {
	log.Logger
	store         *store.Store
	client        client.Client
	styles        styles
	clientTimeout time.Duration // how long to wait for a response
	clientContext int           // how much chat context to send
}

func (uiOpts uiOpts) WithLogPrefix(prefix string) uiOpts {
	uiOpts.Logger = log.Prefixed(prefix, uiOpts.Logger)
	return uiOpts
}

type console struct {
	uiOpts
}

func (t *console) Run(ctx context.Context) error {
	//output := termenv.DefaultOutput()
	// cache detected color values
	//termenv.WithColorCache(true)(output)

	//termbox.Init()

	t.Info("\ngpterm starting...\n")
	model := newControlModel(t.uiOpts)

	// Note that using the alt screen buffer hampers our ability to print
	// output above our program.
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
