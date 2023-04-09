package gptea

import "github.com/collinvandyck/gpterm/lib/store"

type ConfigLoadedMsg struct {
	Config store.Config
	Err    error
}
