package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

// NewTemporalInterceptor creates an interceptor that propagates OpenTelemetry
// trace context through Temporal clients, workflows, and activities.
func NewTemporalInterceptor(options temporalotel.TracerOptions) (interceptor.Interceptor, error) {
	if options.Tracer == nil {
		options.Tracer = GetTracer()
	}

	return temporalotel.NewTracingInterceptor(options)
}

// SpanFromWorkflowContext returns the active span created by the Temporal OTel
// interceptor, or a no-op span when tracing is not installed.
func SpanFromWorkflowContext(ctx workflow.Context) trace.Span {
	span, ok := temporalotel.SpanFromWorkflowContext(ctx)
	if ok {
		return span
	}

	return trace.SpanFromContext(context.Background())
}
