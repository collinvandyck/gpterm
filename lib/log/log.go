package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger interface {
	Log(msg string, args ...any)
	New(args ...any) Logger
}

var Default = New()
var Discard = NewWriter(io.Discard)

type Option func(*logger)

func NewWriter(w io.Writer, args ...any) Logger {
	return &logger{
		out:  w,
		args: args,
	}
}

func New(args ...any) Logger {
	return NewWriter(os.Stdout, args...)
}

func Println(msg string, args ...any) {
	Default.Log(msg, args...)
}

var mut sync.Mutex

type logger struct {
	out  io.Writer
	args []any
}

func (l *logger) New(args ...any) Logger {
	return &logger{
		out:  l.out,
		args: append(l.args, args...),
	}
}

func (l *logger) Log(msg string, args ...any) {
	l.println(l.out, msg, args...)
}

func (l *logger) println(w io.Writer, msg string, args ...any) {
	rendered := l.render(msg, args...)
	mut.Lock()
	defer mut.Unlock()
	fmt.Fprintln(w, rendered)
}

func (l *logger) render(msg string, args ...any) string {
	buf := new(bytes.Buffer)
	l.renderPair(buf, "t", time.Now())
	l.renderPair(buf, "msg", msg)
	l.renderArgs(buf, l.args)
	l.renderArgs(buf, args)
	return buf.String()
}

func (l *logger) renderArgs(buf *bytes.Buffer, args []any) {
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		ks, ok := key.(string)
		if !ok {
			ks = fmt.Sprintf("%v", key)
		}
		l.renderPair(buf, ks, val)
	}
	if len(args)%2 == 1 {
		l.renderPair(buf, "!!odd!!", args[len(args)-1])
	}
}

func (l *logger) renderPair(buf *bytes.Buffer, key string, val any) {
	if buf.Len() > 0 {
		buf.WriteString(" ")
	}
	buf.WriteString(key)
	buf.WriteString("=")
	quote := func(str string) string {
		return "\"" + str + "\""
	}
	maybeQuote := func(str string) string {
		// check if we need to quote. if str has spaces, or is empty, or is a number, quote it.
		if strings.ContainsAny(str, " \t\n") {
			str = quote(str)
		}
		return str
	}
	switch v := val.(type) {
	case string:
		if key == "msg" {
			buf.WriteString(quote(v))
		} else {
			buf.WriteString(maybeQuote(v))
		}
	case []byte:
		buf.WriteString(maybeQuote(string(v)))
	case time.Time:
		ft := v.Format("15:04:05")
		buf.WriteString(maybeQuote(ft))
	default:
		fmt.Fprint(buf, v)
	}

}
