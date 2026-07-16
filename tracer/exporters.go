package tracer

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type TracerTransport string

const (
	TransportJaegerHTTP TracerTransport = "jaeger/http"
	TransportJaegerUDP  TracerTransport = "jaeger/udp"
	TransportOTLPGRPC   TracerTransport = "otel/grpc"
	TransportOTLPHTTP   TracerTransport = "otel/http"
)

type JaegerTracerOptions struct {
	Sampler   sdktrace.Sampler
	Host      string
	Port      string
	Transport TracerTransport
}

// SetupJaegerTracerProviderHTTP configures the legacy direct Jaeger exporter.
// Deprecated: use SetupOTLPGRPC or SetupOTLPHTTP with an OpenTelemetry Collector.
func SetupJaegerTracerProviderHTTP(
	ctx context.Context,
	o JaegerTracerOptions,
	namespace, serviceName, env string,
) error {
	if namespace != "" {
		serviceName = namespace + "." + serviceName
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return err
	}

	var exporter sdktrace.SpanExporter

	// Set up a trace exporter
	switch o.Transport {
	case TransportJaegerHTTP:
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(fmt.Sprintf("http://%s:%s/api/traces", o.Host, o.Port))),
		)
	case TransportJaegerUDP:
		exporter, err = jaeger.New(
			jaeger.WithAgentEndpoint(jaeger.WithAgentHost(o.Host), jaeger.WithAgentPort(o.Port)),
		)
	case TransportOTLPGRPC:
		exporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", o.Host, o.Port)),
			otlptracegrpc.WithInsecure(),
		))
	case TransportOTLPHTTP:
		exporter, err = otlptrace.New(ctx, otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%s", o.Host, o.Port)),
			otlptracehttp.WithInsecure(),
		))
	default:
		return fmt.Errorf(
			"unexpected type of exporter: %s, allowed types: %v",
			o.Transport,
			[]TracerTransport{
				TransportJaegerHTTP,
				TransportJaegerUDP,
				TransportOTLPGRPC,
				TransportOTLPHTTP,
			},
		)
	}

	if err != nil {
		return errors.New("failed to create trace exporter")
	}

	traceProviderOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(o.Sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	}

	tracerProvider := sdktrace.NewTracerProvider(
		traceProviderOpts...,
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	SetupGlobalTracer(otel.Tracer(serviceName), serviceName)

	return nil
}
