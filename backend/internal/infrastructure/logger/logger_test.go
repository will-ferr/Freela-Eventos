package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("log não é JSON válido: %v (%s)", err, buf.String())
	}
	return m
}

func TestLogger_AddsServiceAndRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "info", "json", "workevent-api", "test")

	ctx := WithRequestID(context.Background(), "req-123")
	log.InfoContext(ctx, "olá")

	m := decode(t, &buf)
	if m["service"] != "workevent-api" || m["env"] != "test" {
		t.Errorf("atributos fixos ausentes: %v", m)
	}
	if m["request_id"] != "req-123" {
		t.Errorf("request_id ausente: %v", m)
	}
	if _, ok := m["trace_id"]; ok {
		t.Error("trace_id não deveria existir sem span ativo")
	}
}

func TestLogger_AddsTraceIDWhenSpanPresent(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "info", "json", "svc", "test")

	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	log.InfoContext(ctx, "com trace")

	m := decode(t, &buf)
	if m["trace_id"] != traceID.String() {
		t.Errorf("trace_id = %v", m["trace_id"])
	}
}

func TestLogger_RespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "warn", "json", "svc", "test")

	log.Info("não deve aparecer")
	if buf.Len() != 0 {
		t.Errorf("info não deveria ser registrado com nível warn: %s", buf.String())
	}
	log.Warn("deve aparecer")
	if buf.Len() == 0 {
		t.Error("warn deveria ser registrado")
	}
}
