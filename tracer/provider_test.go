package tracer

import (
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewProviderRequiresServiceNameAndExporter(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	if _, err := NewProvider(t.Context(), Config{}, exporter); err == nil {
		t.Error("NewProvider() error = nil, want service name validation error")
	}
	if _, err := NewProvider(t.Context(), Config{ServiceName: "worker"}, nil); err == nil {
		t.Error("NewProvider() error = nil, want exporter validation error")
	}
}

func TestNewProviderFlushesAndShutsDown(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider, err := NewProvider(t.Context(), Config{
		ServiceName: "worker",
		Sampler:     sdktrace.AlwaysSample(),
	}, exporter)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	_, span := provider.Tracer("test").Start(t.Context(), "run")
	span.End()
	if err := provider.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v", err)
	}
	if got := len(exporter.GetSpans()); got != 1 {
		t.Errorf("exported spans = %d, want 1", got)
	}
	if err := provider.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
