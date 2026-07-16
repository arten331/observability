package tracer

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewProviderRequiresServiceNameAndExporter(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	if _, err := NewProvider(context.Background(), Config{}, exporter); err == nil {
		t.Error("NewProvider() error = nil, want service name validation error")
	}
	if _, err := NewProvider(context.Background(), Config{ServiceName: "worker"}, nil); err == nil {
		t.Error("NewProvider() error = nil, want exporter validation error")
	}
}

func TestNewProviderFlushesAndShutsDown(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider, err := NewProvider(context.Background(), Config{
		ServiceName: "worker",
		Sampler:     sdktrace.AlwaysSample(),
	}, exporter)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	_, span := provider.Tracer("test").Start(context.Background(), "run")
	span.End()
	if err := provider.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush() error = %v", err)
	}
	if got := len(exporter.GetSpans()); got != 1 {
		t.Errorf("exported spans = %d, want 1", got)
	}
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
