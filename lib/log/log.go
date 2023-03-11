package log

import "log"

type Logger interface {
	Info(msg string, args ...any)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

var defaultLogger Logger = console{}

type console struct {
}

func (c console) Info(msg string, args ...any) {
	log.Printf(msg, args...)
}
