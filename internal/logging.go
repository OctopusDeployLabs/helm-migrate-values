package internal

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	IsDebug bool
}

func NewLogger(isDebug bool) *Logger {
	return &Logger{IsDebug: isDebug}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.IsDebug {
		format = fmt.Sprintf("[debug] %s\n", format)
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warning(format string, v ...interface{}) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	_, _ = fmt.Fprintf(os.Stderr, format, v...)
}

func (l *Logger) Information(format string, v ...interface{}) {
	format = fmt.Sprintf("INFO: %s\n", format)
	_, _ = fmt.Fprintf(os.Stdout, format, v...)
}
