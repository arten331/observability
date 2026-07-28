# Observability

Zap-based logging, OpenTelemetry tracing, and Temporal bindings for Go services.

## What is included

- Zap logger with `console`, `json`, and the legacy-compatible `oculus` encoding.
- OTLP tracing over gRPC or HTTP with service resource attributes and an explicit provider shutdown path.
- Temporal client and worker interceptors that propagate OpenTelemetry trace context.
- Oculus field propagation across Temporal workflow and activity boundaries.

Services should export OTLP to an OpenTelemetry Collector.

## Uber Fx initialization

`logger.Module()` receives three named configuration values. `tracer.Module()`
starts OTLP and shuts down the provider with the Fx lifecycle.
`temporal.Module()` dials and closes the Temporal client after tracing is ready.
`metrics.Module()` registers every `metrics.MetricableService` supplied in the
`metricable_services` value group.

Provide the configuration values in the composition root, then include the
modules in the Fx app:

```go
package main

import (
	"context"

	"github.com/arten331/observability/logger"
	"github.com/arten331/observability/metrics"
	"github.com/arten331/observability/temporal"
	"github.com/arten331/observability/tracer"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

func NewApp(cfg *Config) *fx.App {
	return fx.New(
		fx.Supply(
			cfg,
			tracer.Config{
				ServiceName: cfg.Tracer.ServiceName,
				Version:     cfg.Version,
				Environment: cfg.Environment,
			},
			tracer.OTLPConfig{
				Protocol: tracer.OTLPGRPC,
				Endpoint: cfg.Tracer.Endpoint,
				Insecure: cfg.Tracer.Insecure,
			},
		),
		fx.Provide(
			fx.Annotate(
				func(cfg *Config) string { return cfg.Logger.Level },
				fx.ResultTags(`name:"log_level"`),
			),
			fx.Annotate(
				func(cfg *Config) string { return cfg.Logger.Output },
				fx.ResultTags(`name:"log_output"`),
			),
			fx.Annotate(
				func(cfg *Config) string { return cfg.Logger.Encoding },
				fx.ResultTags(`name:"log_encoding"`),
			),
			fx.Annotate(
				NewPaymentsMetrics,
				fx.As(new(metrics.MetricableService)),
				fx.ResultTags(`group:"metricable_services"`),
			),
			func(cfg *Config) client.Options {
				return client.Options{HostPort: cfg.Temporal.HostPort}
			},
		),
		logger.Module(),
		tracer.Module(),
		temporal.Module(),
		metrics.Module(),
		fx.Invoke(registerTemporalWorker),
	)
}

func registerTemporalWorker(
	lifecycle fx.Lifecycle,
	temporalClient client.Client,
	cfg *Config,
) {
	temporalWorker := worker.New(temporalClient, cfg.Temporal.TaskQueue, worker.Options{})
	// temporalWorker.RegisterWorkflow(YourWorkflow)
	// temporalWorker.RegisterActivity(YourActivities)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error { return temporalWorker.Start() },
		OnStop: func(context.Context) error {
			temporalWorker.Stop()
			return nil
		},
	})
}

func main() {
	NewApp(loadConfig()).Run()
}
```

`NewPaymentsMetrics` is an application constructor that returns a value
implementing `metrics.MetricableService`. Add providers for other configuration
sections (for example, `func(cfg *Config) *MigrateOptions { return &cfg.Goose }`)
to the same `fx.Provide` block. The application owns workflow/activity
registration; it starts and stops its worker through the lifecycle hook.

The HTTP and Kafka helpers do not need separate Fx modules: they use the global
OpenTelemetry provider and propagator configured by `tracer.Module()`.

## OpenTelemetry startup without Fx

```go
provider, err := tracer.SetupOTLPGRPC(
	ctx,
	tracer.Config{
		ServiceName: "payments-worker",
		Version:     buildVersion,
		Environment: environment,
	},
	otlptracegrpc.WithEndpoint("otel-collector:4317"),
	otlptracegrpc.WithInsecure(),
)
if err != nil {
	return err
}
defer provider.Shutdown(ctx)
```

## Temporal client and worker without Fx

```go
clientOptions, err := temporal.WithClientObservability(client.Options{})
if err != nil {
	return err
}
temporalClient, err := client.Dial(clientOptions)
if err != nil {
	return err
}

workerOptions, err := temporal.WithWorkerObservability(worker.Options{})
if err != nil {
	return err
}
temporalWorker := worker.New(temporalClient, "payments", workerOptions)
```

Use `logger.TraceWithID` before starting a workflow, or `logger.TraceWFWithID` inside a workflow, to carry an Oculus field to the following workflow and activity calls.
