package util

import (
	"encoding/json"
	"fmt"
	"io"
)

type logger struct {
	w io.Writer
}

type Logger interface {
	Log(level LogLevel, msg string, vars ...LogEnv)
}

func NewLogger(w io.Writer) Logger {
	return &logger{w}
}

type LogEnv struct {
	Key   string
	Value any
}

type LogLevel uint

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

func (l LogLevel) String() string {
	switch l {
	case Trace:
		return "Trace"
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Warn:
		return "Warn"
	case Error:
		return "Error"
	case Fatal:
		return "Fatal"
	default:
		return "Unknown"
	}
}

func (l *logger) Log(level LogLevel, msg string, vars ...LogEnv) {
	logMsg := struct {
		msg  string
		vars []LogEnv
	}{
		msg:  fmt.Sprintf("[%s]: %s", level.String(), msg),
		vars: vars,
	}

	logJson, _ := json.Marshal(logMsg)

	fmt.Fprintln(l.w, string(logJson))
}
