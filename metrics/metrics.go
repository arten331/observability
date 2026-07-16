package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricableService interface {
	GetMetrics() []prometheus.Collector
}

type PrometheusService interface {
	Register(prometheus.Collector) error
	AddMiddleware(func(handler http.Handler) http.Handler)
}

type Service struct {
	registry *prometheus.Registry
	handler  http.HandlerFunc
}

func (p *Service) AddMiddleware(mw func(next http.Handler) http.Handler) {
	p.handler = mw(p.handler).ServeHTTP
}

func New() Service {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())

	return Service{
		registry: reg,
		handler: func(w http.ResponseWriter, r *http.Request) {
			promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		},
	}
}

func (p Service) Handler() http.HandlerFunc {
	return p.handler
}

func (p Service) Register(c prometheus.Collector) error {
	return p.registry.Register(c)
}

func (p Service) RegisterService(s MetricableService) error {
	metrics := s.GetMetrics()
	for _, metric := range metrics {
		err := p.registry.Register(metric)
		if err != nil {
			return err
		}
	}

	return nil
}
