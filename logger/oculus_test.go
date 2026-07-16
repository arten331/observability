package logger

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestOculusEncoding(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp(t.TempDir(), "oculus-*.log")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	log, err := New(WithConfiguration(CoreOptions{
		OutputPath: file.Name(),
		Level:      KeyLevelInfo,
		Encoding:   EncodingOculus,
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("workflow started", zap.String("workflow_id", "wf-42"))
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(content); !strings.Contains(got, `"text":"workflow started"`) {
		t.Errorf("log = %q, want Oculus text field", got)
	}
	if got := string(content); !strings.Contains(got, `"workflow_id":"wf-42"`) {
		t.Errorf("log = %q, want structured field", got)
	}
}
