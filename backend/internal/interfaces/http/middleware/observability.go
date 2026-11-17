package middleware

import (
	"log/slog"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/telemetry"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

// routeLabel usa o template da rota (/users/:id) e não o path real, para que
// a cardinalidade das métricas não cresça com IDs.
func routeLabel(c *gin.Context) string {
	if r := c.FullPath(); r != "" {
		return r
	}
	return "unmatched"
}

// AccessLog registra uma linha estruturada por requisição. Rotas de alta
// frequência e baixo valor (probes, scrape) são registradas em debug.
func AccessLog(log *slog.Logger, quiet ...string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(quiet))
	for _, p := range quiet {
		skip[p] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		if _, ok := skip[c.FullPath()]; ok && status < 400 {
			level = slog.LevelDebug
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("route", routeLabel(c)),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("bytes", c.Writer.Size()),
			slog.String("user_agent", c.Request.UserAgent()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.Last().Error()))
		}
		log.LogAttrs(c.Request.Context(), level, "http request", attrs...)
	}
}

// Metrics alimenta os contadores e histogramas RED (Rate, Errors, Duration).
func Metrics(m *telemetry.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.InFlight.Inc()
		defer m.InFlight.Dec()

		c.Next()

		route := routeLabel(c)
		m.RequestsTotal.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.RequestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

// Recovery transforma panics em 500 padronizado, registra a stack trace e
// incrementa a métrica de panics (que alimenta um alerta no Prometheus).
func Recovery(log *slog.Logger, m *telemetry.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				if m != nil {
					m.Panics.Inc()
				}
				log.ErrorContext(c.Request.Context(), "panic recuperado",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				response.AbortWith(c, 500, "INTERNAL_ERROR",
					"Ocorreu um erro inesperado. Tente novamente em instantes.")
			}
		}()
		c.Next()
	}
}
