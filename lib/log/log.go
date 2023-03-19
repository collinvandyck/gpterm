package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

var Default = New()
var Discard = New(WithStdout(io.Discard), WithStderr(io.Discard))

type Option func(*logger)

func WithStdout(w io.Writer) Option {
	return func(l *logger) {
		l.out = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(l *logger) {
		l.err = w
	}
}

func New(opts ...Option) Logger {
	res := &logger{
		out: os.Stdout,
		err: os.Stderr,
	}
	for _, o := range opts {
		o(res)
	}
	return res
}

func Info(msg string, args ...any) {
	Default.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Default.Error(msg, args...)
}

type logger struct {
	out io.Writer
	err io.Writer
	mut sync.Mutex
}

func (l *logger) Info(msg string, args ...any) {
	fmt.Fprintf(l.out, msg+"\n", args...)
}

func (l *logger) Error(msg string, args ...any) {
	fmt.Fprintf(l.err, msg+"|n", args...)
}

func (l *logger) print(w io.Writer, msg string, args ...any) {
	l.mut.Lock()
	defer l.mut.Unlock()
	fmt.Fprintf(w, msg+"\n", args...)
}
