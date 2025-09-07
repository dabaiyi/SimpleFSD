// Package log
package log

import "github.com/half-nothing/simple-fsd/internal/interfaces/global"

type LoggerInterface interface {
	Init(debug bool)
	ShutdownCallback() global.Callable
	Debug(msg string, v ...interface{})
	DebugF(msg string, v ...interface{})
	Info(msg string, v ...interface{})
	InfoF(msg string, v ...interface{})
	Warn(msg string, v ...interface{})
	WarnF(msg string, v ...interface{})
	Error(msg string, v ...interface{})
	ErrorF(msg string, v ...interface{})
	Fatal(msg string, v ...interface{})
	FatalF(msg string, v ...interface{})
}
