package util

import (
	"encoding/json"
	"fmt"
	"log"
)

type logger struct {
	logger *log.Logger
}

type Logger interface {
	Log(level LogLevel, msg string, vars ...LogEnv)
}

func NewLogger(l *log.Logger) Logger {
	return &logger{l}
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
		Msg  string
		Vars []LogEnv
	}{
		Msg:  fmt.Sprintf("[%s]: %s", level.String(), msg),
		Vars: vars,
	}
	logJson, _ := json.Marshal(logMsg)

	l.logger.Print(string(logJson))
}
