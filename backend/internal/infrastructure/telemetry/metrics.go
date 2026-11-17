// Package telemetry concentra métricas (Prometheus) e tracing (OpenTelemetry).
package telemetry

import (
	"database/sql"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "workevent"

// Metrics usa um registry próprio (e não o global) para que testes possam
// criar instâncias isoladas sem colisão de registro.
type Metrics struct {
	Registry        *prometheus.Registry
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	InFlight        prometheus.Gauge
	RateLimited     prometheus.Counter
	Panics          prometheus.Counter
}

func NewMetrics(build map[string]string) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		Registry: reg,
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace, Subsystem: "http", Name: "requests_total",
			Help: "Total de requisições HTTP por método, rota e status.",
		}, []string{"method", "route", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace, Subsystem: "http", Name: "request_duration_seconds",
			Help:    "Latência das requisições HTTP.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		}, []string{"method", "route"}),
		InFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Subsystem: "http", Name: "requests_in_flight",
			Help: "Requisições sendo processadas no momento.",
		}),
		RateLimited: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Subsystem: "http", Name: "rate_limited_total",
			Help: "Requisições rejeitadas pelo rate limiter.",
		}),
		Panics: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Subsystem: "http", Name: "panics_recovered_total",
			Help: "Panics recuperados pelo middleware de recovery.",
		}),
	}

	buildInfo := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace, Name: "build_info",
		Help:        "Versão em execução (valor sempre 1).",
		ConstLabels: build,
	})
	buildInfo.Set(1)

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.RequestsTotal, m.RequestDuration, m.InFlight, m.RateLimited, m.Panics,
		buildInfo,
	)
	return m
}

// RegisterDB expõe estatísticas do pool de conexões (abertas, em uso, espera).
func (m *Metrics) RegisterDB(db *sql.DB, name string) {
	m.Registry.MustRegister(collectors.NewDBStatsCollector(db, name))
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{Registry: m.Registry})
}
