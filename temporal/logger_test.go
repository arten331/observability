package temporal

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerWarnUsesWarnLevelAndStructuredFields(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.DebugLevel)
	log := NewLogger(zap.New(core))
	log.Warn("task retry", "attempt", 2)

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if got, want := entries[0].Level, zap.WarnLevel; got != want {
		t.Errorf("level = %s, want %s", got, want)
	}
	if got, want := entries[0].ContextMap()["attempt"], int64(2); got != want {
		t.Errorf("attempt = %v, want %d", got, want)
	}
}
