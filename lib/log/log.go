package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

var Default = New()
var Discard = New(WithStdout(io.Discard), WithStderr(io.Discard))

var maxPrefixLen int64
var maxPrefixLenMut sync.Mutex

type Option func(*logger)

func WithWriter(w io.Writer) Option {
	return func(l *logger) {
		l.out = w
		l.err = w
	}
}

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

func Prefixed(prefix string, log Logger) Logger {
	prefix = "[" + prefix + "]"
	maxPrefixLenMut.Lock()
	defer maxPrefixLenMut.Unlock()
	if len(prefix) > int(maxPrefixLen) {
		maxPrefixLen = int64(len(prefix))
	}
	return prefixedLogger{
		Logger: log,
		prefix: prefix,
	}
}

type prefixedLogger struct {
	Logger
	prefix string
}

func (p prefixedLogger) Info(msg string, args ...any) {
	p.Logger.Info(p.padPrefix()+" "+msg, args...)
}
func (p prefixedLogger) Error(msg string, args ...any) {
	p.Logger.Error(p.padPrefix()+" "+msg, args...)
}

func (p prefixedLogger) padPrefix() string {
	prefix := p.prefix
	ml := int(atomic.LoadInt64(&maxPrefixLen))
	if len(prefix) < ml {
		prefix = prefix + strings.Repeat(" ", ml-len(p.prefix))
	}
	return prefix
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
	out       io.Writer
	err       io.Writer
	mut       sync.Mutex
	printTime bool
}

func (l *logger) Info(msg string, args ...any) {
	l.print(l.out, msg, args...)
}

func (l *logger) Error(msg string, args ...any) {
	l.print(l.err, msg, args...)
}

func (l *logger) print(w io.Writer, msg string, args ...any) {
	l.mut.Lock()
	defer l.mut.Unlock()
	if l.printTime {
		now := time.Now().Unix()
		msg = fmt.Sprintf("%d %s", now, msg)
	}
	fmt.Fprintf(w, msg+"\n", args...)
}
