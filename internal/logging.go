package internal

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	isDebug bool
}

func NewLogger(isDebug bool) *Logger {
	return &Logger{isDebug: isDebug}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.isDebug {
		format = fmt.Sprintf("[debug] %s\n", format)
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warning(format string, v ...interface{}) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	_, _ = fmt.Fprintf(os.Stderr, format, v...)
}
