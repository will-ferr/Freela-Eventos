package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/will-ferr/Freela-Eventos/backend/config"
	"github.com/will-ferr/Freela-Eventos/backend/internal/application/health"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/telemetry"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/middleware"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

type stubHealth struct{ status health.Status }

func (s stubHealth) Readiness(context.Context) health.Report {
	return health.Report{Status: s.status, Checks: map[string]health.CheckResult{"postgres": {Status: s.status}}}
}

func newTestRouter(t *testing.T, status health.Status, mutate func(*config.Config)) nethttp.Handler {
	t.Helper()
	cfg, err := config.Load(config.LookupFromMap(map[string]string{
		"APP_ENV": "test", "DB_PASSWORD": "x",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(cfg)
	}
	r, err := NewRouter(Deps{
		Config:      cfg,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Metrics:     telemetry.NewMetrics(nil),
		Health:      stubHealth{status: status},
		RateLimiter: middleware.NewIPRateLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst),
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func get(h nethttp.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(nethttp.MethodGet, path, nil))
	return w
}

func TestHealthEndpoints(t *testing.T) {
	up := newTestRouter(t, health.StatusUp, nil)
	down := newTestRouter(t, health.StatusDown, nil)

	if w := get(up, "/api/v1/health/live"); w.Code != 200 || !strings.Contains(w.Body.String(), `"ok"`) {
		t.Errorf("live: %d %s", w.Code, w.Body.String())
	}
	if w := get(up, "/api/v1/health/ready"); w.Code != 200 {
		t.Errorf("ready (up) = %d", w.Code)
	}
	if w := get(down, "/api/v1/health/ready"); w.Code != nethttp.StatusServiceUnavailable {
		t.Errorf("ready (down) deveria ser 503, obteve %d", w.Code)
	}
	if w := get(down, "/api/v1/health/live"); w.Code != 200 {
		t.Errorf("live não pode depender do banco, obteve %d", w.Code)
	}
}

func TestEveryResponseHasRequestID(t *testing.T) {
	r := newTestRouter(t, health.StatusUp, nil)
	for _, p := range []string{"/api/v1/health/live", "/rota-inexistente"} {
		if get(r, p).Header().Get(response.HeaderRequestID) == "" {
			t.Errorf("%s sem X-Request-ID", p)
		}
	}
}

func TestNotFoundAndMethodNotAllowedUseStandardBody(t *testing.T) {
	r := newTestRouter(t, health.StatusUp, nil)

	w := get(r, "/api/v1/nada")
	var body response.ErrorBody
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if w.Code != 404 || body.Error.Code != "ROUTE_NOT_FOUND" {
		t.Errorf("404 fora do padrão: %d %s", w.Code, w.Body.String())
	}

	req := httptest.NewRequest(nethttp.MethodDelete, "/api/v1/health/live", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != nethttp.StatusMethodNotAllowed || !strings.Contains(w.Body.String(), "METHOD_NOT_ALLOWED") {
		t.Errorf("405 fora do padrão: %d %s", w.Code, w.Body.String())
	}
}

func TestMetricsEndpointExposesHTTPMetrics(t *testing.T) {
	r := newTestRouter(t, health.StatusUp, nil)
	get(r, "/api/v1/health/live")

	body := get(r, "/metrics").Body.String()
	for _, want := range []string{
		"workevent_http_requests_total",
		`route="/api/v1/health/live"`,
		"workevent_http_request_duration_seconds_bucket",
		"go_goroutines",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics não contém %q", want)
		}
	}
}

func TestRateLimitAppliesOnlyToAPI(t *testing.T) {
	r := newTestRouter(t, health.StatusUp, func(c *config.Config) {
		c.RateLimit.RPS, c.RateLimit.Burst = 0.001, 1
	})

	if get(r, "/api/v1/health/live").Code != 200 {
		t.Fatal("primeira requisição deveria passar")
	}
	if code := get(r, "/api/v1/health/live").Code; code != nethttp.StatusTooManyRequests {
		t.Errorf("segunda requisição deveria ser limitada, obteve %d", code)
	}
	if get(r, "/metrics").Code != 200 {
		t.Error("/metrics (scrape interno) não deve sofrer rate limit")
	}
}

func TestSwaggerCanBeDisabled(t *testing.T) {
	on := newTestRouter(t, health.StatusUp, nil)
	off := newTestRouter(t, health.StatusUp, func(c *config.Config) { c.Swagger = false })

	if code := get(on, "/api/v1/docs/doc.json").Code; code != 200 {
		t.Errorf("swagger habilitado deveria responder 200, obteve %d", code)
	}
	if code := get(off, "/api/v1/docs/doc.json").Code; code != 404 {
		t.Errorf("swagger desabilitado deveria responder 404, obteve %d", code)
	}
}
