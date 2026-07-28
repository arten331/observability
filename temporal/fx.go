package temporal

import (
	"go.opentelemetry.io/otel/sdk/trace"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
)

type clientParams struct {
	fx.In

	Options  client.Options
	Provider *trace.TracerProvider `optional:"true"`
}

// Module dials a Temporal client with Oculus and OpenTelemetry bindings, then
// closes it during Fx shutdown. Include tracer.Module to initialize tracing
// before the client is constructed.
func Module() fx.Option {
	return fx.Module(
		"temporal",
		fx.Provide(newModuleClient),
		fx.Invoke(registerClientClose),
	)
}

func newModuleClient(params clientParams) (client.Client, error) {
	return NewClient(params.Options)
}

func registerClientClose(lifecycle fx.Lifecycle, temporalClient client.Client) {
	lifecycle.Append(fx.StopHook(temporalClient.Close))
}
