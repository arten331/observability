package tracer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer      = trace.NewNoopTracerProvider().Tracer("")
	serviceName = "unknown"
)

func SetupGlobalTracer(t trace.Tracer, sn string) {
	serviceName = sn
	tracer = t
}

func GetTracer() trace.Tracer {
	return tracer
}

func GetServiceName() string {
	return serviceName
}

func NewSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := GetTracer()
	return tracer.Start(ctx, name)
}

func NewSpanWithAttributes(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	var span trace.Span

	tracer := GetTracer()
	ctx, span = tracer.Start(ctx, name)
	span.SetAttributes(attrs...)

	return ctx, span
}

func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
