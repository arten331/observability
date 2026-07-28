package metrics

import (
	"testing"

	"github.com/arten331/observability/logger"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type metricableServiceStub struct {
	collector prometheus.Collector
}

func (s *metricableServiceStub) GetMetrics() []prometheus.Collector {
	return []prometheus.Collector{s.collector}
}

func (s *metricableServiceStub) ModuleName() string {
	return "stub"
}

func TestModuleRegistersMetricableServices(t *testing.T) {
	stub := &metricableServiceStub{
		collector: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "fx_module_test_total",
			Help: "Metric used to verify Fx registration.",
		}),
	}

	var registry MetricsRegistry
	var handler MetricsHandler

	app := fx.New(
		fx.NopLogger,
		logger.Module(),
		Module(),
		fx.Provide(
			fx.Annotate(
				func() string { return logger.KeyLevelDebug },
				fx.ResultTags(`name:"log_level"`),
			),
			fx.Annotate(
				func() string { return "stderr" },
				fx.ResultTags(`name:"log_output"`),
			),
			fx.Annotate(
				func() string { return logger.EncodingConsole },
				fx.ResultTags(`name:"log_encoding"`),
			),
			fx.Annotate(
				func() *metricableServiceStub { return stub },
				fx.As(new(MetricableService)),
				fx.ResultTags(`group:"metricable_services"`),
			),
		),
		fx.Populate(&registry, &handler),
	)

	if err := app.Start(t.Context()); err != nil {
		t.Fatalf("start Fx app: %v", err)
	}
	defer func() {
		if err := app.Stop(t.Context()); err != nil {
			t.Errorf("stop Fx app: %v", err)
		}
	}()

	if registry == nil {
		t.Fatal("Module() did not provide MetricsRegistry")
	}
	if handler == nil || handler.Handler() == nil {
		t.Fatal("Module() did not provide MetricsHandler")
	}

	if err := registry.Register(stub.collector); err == nil {
		t.Fatal("metricable service collector was not registered on start")
	}
}

func TestNewWorksWithoutFx(t *testing.T) {
	service := New()
	collector := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "without_fx_test_total",
		Help: "Metric used to verify direct construction.",
	})

	if err := service.RegisterServices([]MetricableService{
		&metricableServiceStub{collector: collector},
	}); err != nil {
		t.Fatalf("RegisterServices() error = %v", err)
	}
}
