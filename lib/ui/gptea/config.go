package gptea

import (
	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/store"
)

type ConfigLoadedMsg struct {
	Config       store.Config
	ClientConfig query.ClientConfig
	Err          error
}
