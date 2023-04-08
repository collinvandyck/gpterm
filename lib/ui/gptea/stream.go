package gptea

import (
	"context"
	"math/rand"
	"regexp"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type StreamCompletionReq struct {
	Text string
}

type StreamCompletion struct {
	text chan string
	err  chan error
	done chan any
}

type StreamCompletionResult struct {
	Text string
	Err  error
}

func NewStreamCompletion() StreamCompletion {
	return StreamCompletion{
		text: make(chan string, 4096),
		err:  make(chan error, 1),
		done: make(chan any),
	}
}

// Close closes the stream with the error result. This should
// only ever be called once and should always be called.
func (s StreamCompletion) Close(err error) {
	s.err <- err
	close(s.text)
	close(s.done)
}

func (s StreamCompletion) Write(ctx context.Context, text string) error {
	select {
	case s.text <- text:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s StreamCompletion) Next() (res []string) {
	read := true
	max := 100
	for read {
		select {
		case t, open := <-s.text:
			if !open {
				read = false
				break
			}
			res = append(res, t)
			if len(res) == max {
				return
			}
		default:
			read = false
		}
	}
	return
}

func (s StreamCompletion) Err() error {
	select {
	case err := <-s.err:
		s.err <- err
		return err
	default:
		return nil
	}
}

func (s StreamCompletion) Done() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func StaticStream(text string) tea.Cmd {
	return func() tea.Msg {
		csm := NewStreamCompletion()
		go func() {
			defer close(csm.text)
			defer close(csm.done)
			content := make(chan string)
			go func() {
				defer close(content)
				pattern := regexp.MustCompile(`\S+|\s+`)
				fields := pattern.FindAllString(text, -1)
				for _, field := range fields {
					content <- field
					time.Sleep(time.Duration(rand.Int()%25) * time.Millisecond)
				}
			}()

			for c := range content {
				csm.text <- c
			}
		}()
		return csm
	}
}
