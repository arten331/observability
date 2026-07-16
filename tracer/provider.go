package tracer

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Config describes resource attributes and sampling for a tracer provider.
type Config struct {
	ServiceName string
	Version     string
	Environment string
	Sampler     sdktrace.Sampler
}

// NewProvider creates a tracer provider. The caller must eventually call
// Shutdown on the returned provider to flush pending spans.
func NewProvider(ctx context.Context, config Config, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	if config.ServiceName == "" {
		return nil, errors.New("tracer service name is required")
	}
	if exporter == nil {
		return nil, errors.New("tracer exporter is required")
	}

	attributes := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName(config.ServiceName)),
	}
	if config.Version != "" {
		attributes = append(attributes, resource.WithAttributes(semconv.ServiceVersion(config.Version)))
	}
	if config.Environment != "" {
		attributes = append(attributes, resource.WithAttributes(semconv.DeploymentEnvironmentName(config.Environment)))
	}

	res, err := resource.New(ctx, attributes...)
	if err != nil {
		return nil, err
	}

	sampler := config.Sampler
	if sampler == nil {
		sampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
	), nil
}

// SetupOTLPGRPC configures the global provider for an OTLP/gRPC collector.
// The returned provider must be shut down during process termination.
func SetupOTLPGRPC(
	ctx context.Context,
	config Config,
	options ...otlptracegrpc.Option,
) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	provider, err := NewProvider(ctx, config, exporter)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, err
	}

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	SetupGlobalTracer(provider.Tracer(config.ServiceName), config.ServiceName)

	return provider, nil
}

// SetupOTLPHTTP configures the global provider for an OTLP/HTTP collector.
// The returned provider must be shut down during process termination.
func SetupOTLPHTTP(
	ctx context.Context,
	config Config,
	options ...otlptracehttp.Option,
) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	provider, err := NewProvider(ctx, config, exporter)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		return nil, err
	}

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	SetupGlobalTracer(provider.Tracer(config.ServiceName), config.ServiceName)

	return provider, nil
}
