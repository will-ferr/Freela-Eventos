package telemetry

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

// Regressão: o merge do resource com semconv fixo falhava em runtime
// ("conflicting Schema URL") apenas com tracing habilitado.
func TestSetupTracing_EnabledInitializesProvider(t *testing.T) {
	ctx := context.Background()
	shutdown, err := SetupTracing(ctx, TracingOptions{
		Enabled:     true,
		ServiceName: "test",
		Version:     "1.0.0",
		Env:         "test",
		Endpoint:    "127.0.0.1:1", // exportação é assíncrona; não precisa estar no ar
		SampleRatio: 1,
	})
	if err != nil {
		t.Fatalf("SetupTracing falhou: %v", err)
	}

	_, span := otel.Tracer("t").Start(ctx, "op")
	if !span.SpanContext().IsValid() {
		t.Error("com tracing habilitado o span deveria ser real (não no-op)")
	}
	span.End()

	sctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	_ = shutdown(sctx) // o exportador pode falhar ao enviar; o importante é não travar
}

func TestSetupTracing_DisabledIsNoop(t *testing.T) {
	shutdown, err := SetupTracing(context.Background(), TracingOptions{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown no-op não deveria falhar: %v", err)
	}
}

func TestMetrics_ExposesBuildInfoAndDBStats(t *testing.T) {
	m := NewMetrics(map[string]string{"version": "1.2.3"})
	db := sql.OpenDB(fakeConnector{}) // não abre conexão real até ser usado
	m.RegisterDB(db, "workevent")

	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	body, _ := io.ReadAll(w.Body)

	for _, want := range []string{`workevent_build_info{version="1.2.3"} 1`, `go_sql_max_open_connections{db_name="workevent"}`} {
		if !strings.Contains(string(body), want) {
			t.Errorf("métricas não contêm %q", want)
		}
	}
}

type fakeConnector struct{}

func (fakeConnector) Connect(context.Context) (driver.Conn, error) { return nil, driver.ErrBadConn }
func (fakeConnector) Driver() driver.Driver                        { return nil }
