package metrics

import (
	"fmt"

	"github.com/arten331/observability/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RegistryParams struct {
	fx.In

	Logger   *logger.Logger
	Registry MetricsRegistry
	Services MetricableServices
}

type MetricsRegistryParams struct {
	fx.In

	Services []MetricableService `group:"metricable_services"`
}

// MetricableServices wraps the Fx value group so it can be consumed as one
// dependency by the registry initializer.
type MetricableServices struct {
	Services []MetricableService
}

// Module provides the Prometheus registry and handler and registers all
// MetricableService values from the "metricable_services" Fx value group.
// The regular New constructor remains available for applications without Fx.
func Module() fx.Option {
	return fx.Module(
		"metrics_registry",
		fx.Provide(
			fx.Annotate(
				New,
				fx.As(new(MetricsRegistry)),
				fx.As(new(MetricsHandler)),
			),
			func(params MetricsRegistryParams) MetricableServices {
				return MetricableServices{Services: params.Services}
			},
		),
		fx.Invoke(initRegisterMetrics),
	)
}

func initRegisterMetrics(params RegistryParams) error {
	if err := params.Registry.RegisterServices(params.Services.Services); err != nil {
		return err
	}

	for _, service := range params.Services.Services {
		params.Logger.Info(
			"registered metrics",
			zap.String("module", metricableServiceName(service)),
		)
	}

	return nil
}

func metricableServiceName(service MetricableService) string {
	if named, ok := service.(interface{ ModuleName() string }); ok {
		return named.ModuleName()
	}

	return fmt.Sprintf("%T", service)
}
