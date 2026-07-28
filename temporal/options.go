package temporal

import (
	"github.com/arten331/observability/logger"
	"github.com/arten331/observability/tracer"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

// WithClientObservability adds Oculus logging, persistent field propagation,
// and OpenTelemetry tracing to Temporal client options.
func WithClientObservability(options client.Options) (client.Options, error) {
	tracing, err := newTracingInterceptor()
	if err != nil {
		return options, err
	}

	options.Logger = NewGlobalLogger()
	options.ContextPropagators = append(options.ContextPropagators, logger.NewPersistenceValuesPropagator())
	options.Interceptors = append(options.Interceptors, tracing)

	return options, nil
}

// WithWorkerObservability adds OpenTelemetry tracing to Temporal worker options.
// Use WithClientObservability for the client that constructs the worker too, so
// persisted Oculus fields are available across workflow boundaries.
func WithWorkerObservability(options worker.Options) (worker.Options, error) {
	tracing, err := newTracingInterceptor()
	if err != nil {
		return options, err
	}

	options.Interceptors = append(options.Interceptors, tracing)
	return options, nil
}

// NewClient dials Temporal with the standard Oculus observability bindings.
func NewClient(options client.Options) (client.Client, error) {
	options, err := WithClientObservability(options)
	if err != nil {
		return nil, err
	}

	clientInstance, err := client.Dial(options)
	if err != nil {
		return nil, logger.L().WithError("failed to create Temporal client", err)
	}

	return clientInstance, nil
}

// NewClientWithLogger is retained for source compatibility. It now enables the
// complete Oculus and OpenTelemetry integration.
func NewClientWithLogger(options client.Options) (client.Client, error) {
	return NewClient(options)
}

func newTracingInterceptor() (interceptor.Interceptor, error) {
	return tracer.NewTemporalInterceptor(temporalotel.TracerOptions{
		Tracer:            tracer.GetTracer(),
		TextMapPropagator: otel.GetTextMapPropagator(),
	})
}
