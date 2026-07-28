# Observability

Zap-based logging, OpenTelemetry tracing, and Temporal bindings for Go services.

## What is included

- Zap logger with `console`, `json`, and the legacy-compatible `oculus` encoding.
- OTLP tracing over gRPC or HTTP with service resource attributes and an explicit provider shutdown path.
- Temporal client and worker interceptors that propagate OpenTelemetry trace context.
- Oculus field propagation across Temporal workflow and activity boundaries.

Services should export OTLP to an OpenTelemetry Collector.

## OpenTelemetry startup

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

## Temporal client and worker

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
