// Package logger configura o slog da aplicação e enriquece cada registro com
// request_id e trace_id presentes no contexto, permitindo correlacionar logs,
// métricas e traces de uma mesma requisição.
package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

// WithRequestID grava o request id no contexto para que todo log feito com
// *Context(ctx, ...) o inclua automaticamente.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

// New cria o logger raiz. format: "json" (produção, ingestão por Loki/ELK)
// ou "text" (legível no terminal durante o desenvolvimento).
func New(w io.Writer, level, format, service, env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}

	return slog.New(&contextHandler{Handler: h}).With(
		slog.String("service", service),
		slog.String("env", env),
	)
}

func parseLevel(l string) slog.Level {
	switch strings.ToLower(l) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type contextHandler struct{ slog.Handler }

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestIDFrom(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}
