// Package temporal adapts Oculus logging and tracing to the Temporal Go SDK.
package temporal

import (
	"fmt"

	"github.com/arten331/observability/logger"
	temporallog "go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// Logger adapts a Zap logger to Temporal's key/value logging interface.
type Logger struct {
	log *zap.Logger
}

var _ temporallog.Logger = (*Logger)(nil)

// NewLogger creates a Temporal logger from a Zap logger.
func NewLogger(log *zap.Logger) *Logger {
	if log == nil {
		log = zap.NewNop()
	}

	return &Logger{log: log.WithOptions(zap.AddCallerSkip(1))}
}

// NewGlobalLogger adapts the configured global Oculus logger.
func NewGlobalLogger() *Logger {
	return NewLogger(logger.L().Logger)
}

func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log.Debug(msg, fields(keyvals)...)
}

func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log.Info(msg, fields(keyvals)...)
}

func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log.Warn(msg, fields(keyvals)...)
}

func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log.Error(msg, fields(keyvals)...)
}

func fields(keyvals []interface{}) []zap.Field {
	fields := make([]zap.Field, 0, (len(keyvals)+1)/2)
	for index := 0; index+1 < len(keyvals); index += 2 {
		key, ok := keyvals[index].(string)
		if !ok {
			key = fmt.Sprint(keyvals[index])
		}
		fields = append(fields, zap.Any(key, keyvals[index+1]))
	}
	if len(keyvals)%2 != 0 {
		fields = append(fields, zap.Any("temporal.unpaired_key", keyvals[len(keyvals)-1]))
	}

	return fields
}
