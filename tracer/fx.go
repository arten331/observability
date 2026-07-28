package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

type OTLPProtocol string

const (
	OTLPGRPC OTLPProtocol = "grpc"
	OTLPHTTP OTLPProtocol = "http"
)

// OTLPConfig configures the OTLP exporter used by Module.
type OTLPConfig struct {
	Protocol OTLPProtocol
	Endpoint string
	Insecure bool
}

type moduleParams struct {
	fx.In

	TracerConfig Config
	OTLPConfig   OTLPConfig
}

// Module configures the global OpenTelemetry provider and shuts it down with
// the Fx lifecycle. Applications supply Config and OTLPConfig with fx.Supply
// or their own constructors.
func Module() fx.Option {
	return fx.Module(
		"tracer",
		fx.Provide(newModuleProvider),
		fx.Invoke(registerProviderShutdown),
	)
}

func newModuleProvider(params moduleParams) (*sdktrace.TracerProvider, error) {
	switch params.OTLPConfig.Protocol {
	case "", OTLPGRPC:
		options := make([]otlptracegrpc.Option, 0, 2)
		if params.OTLPConfig.Endpoint != "" {
			options = append(options, otlptracegrpc.WithEndpoint(params.OTLPConfig.Endpoint))
		}
		if params.OTLPConfig.Insecure {
			options = append(options, otlptracegrpc.WithInsecure())
		}

		return SetupOTLPGRPC(context.Background(), params.TracerConfig, options...)
	case OTLPHTTP:
		options := make([]otlptracehttp.Option, 0, 2)
		if params.OTLPConfig.Endpoint != "" {
			options = append(options, otlptracehttp.WithEndpoint(params.OTLPConfig.Endpoint))
		}
		if params.OTLPConfig.Insecure {
			options = append(options, otlptracehttp.WithInsecure())
		}

		return SetupOTLPHTTP(context.Background(), params.TracerConfig, options...)
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol %q", params.OTLPConfig.Protocol)
	}
}

func registerProviderShutdown(lifecycle fx.Lifecycle, provider *sdktrace.TracerProvider) {
	lifecycle.Append(fx.StopHook(provider.Shutdown))
}
