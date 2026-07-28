package tracer

import (
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

func TestModule(t *testing.T) {
	var provider *trace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		Module(),
		fx.Supply(
			Config{ServiceName: "test-service"},
			OTLPConfig{Protocol: OTLPHTTP, Endpoint: "localhost:4318", Insecure: true},
		),
		fx.Populate(&provider),
	)

	if err := app.Start(t.Context()); err != nil {
		t.Fatalf("start Fx app: %v", err)
	}
	defer func() {
		if err := app.Stop(t.Context()); err != nil {
			t.Errorf("stop Fx app: %v", err)
		}
	}()

	if provider == nil {
		t.Fatal("Module() did not provide a tracer provider")
	}
}

func TestModuleRejectsUnsupportedProtocol(t *testing.T) {
	app := fx.New(
		fx.NopLogger,
		Module(),
		fx.Supply(
			Config{ServiceName: "test-service"},
			OTLPConfig{Protocol: "invalid"},
		),
	)

	if err := app.Start(t.Context()); err == nil {
		t.Fatal("start Fx app with unsupported protocol: expected error")
	}
}
