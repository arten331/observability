package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNamedScopesEveryLoggerViewWithoutMutatingParent(t *testing.T) {
	core, observed := observer.New(zap.DebugLevel)
	baseZap := zap.New(core)
	base := &Logger{
		Logger:        baseZap,
		SugaredLogger: baseZap.Sugar(),
		CtxLogger:     baseZap,
		ErrorLogger:   baseZap,
	}

	scoped := base.Named("delivery").Named("temporal")
	scoped.Info("logger")
	scoped.SugaredLogger.Info("sugared")
	scoped.CtxLogger.Info("context")
	scoped.ErrorLogger.Info("error")
	base.Info("parent")

	entries := observed.All()
	if len(entries) != 5 {
		t.Fatalf("log entry count = %d, want 5", len(entries))
	}

	for _, entry := range entries[:4] {
		if entry.LoggerName != "delivery.temporal" {
			t.Errorf("scoped logger name = %q, want %q", entry.LoggerName, "delivery.temporal")
		}
	}

	if entries[4].LoggerName != "" {
		t.Errorf("parent logger name = %q, want empty", entries[4].LoggerName)
	}
}
