package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/logger"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/telemetry"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

func init() { gin.SetMode(gin.TestMode) }

func do(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- RequestID -------------------------------------------------------------

func TestRequestID(t *testing.T) {
	var seenInCtx string
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { seenInCtx = logger.RequestIDFrom(c.Request.Context()) })

	t.Run("gera quando ausente", func(t *testing.T) {
		w := do(r, "GET", "/", nil)
		id := w.Header().Get(response.HeaderRequestID)
		if len(id) != 36 {
			t.Fatalf("esperava UUID, obteve %q", id)
		}
		if seenInCtx != id {
			t.Errorf("id no contexto (%q) difere do header (%q)", seenInCtx, id)
		}
	})

	t.Run("propaga quando válido", func(t *testing.T) {
		w := do(r, "GET", "/", map[string]string{response.HeaderRequestID: "nginx-abc_123.4"})
		if got := w.Header().Get(response.HeaderRequestID); got != "nginx-abc_123.4" {
			t.Errorf("id = %q", got)
		}
	})

	t.Run("substitui quando inseguro", func(t *testing.T) {
		evil := "abc\nlevel=error msg=forjado"
		w := do(r, "GET", "/", map[string]string{response.HeaderRequestID: evil})
		if got := w.Header().Get(response.HeaderRequestID); got == evil || len(got) != 36 {
			t.Errorf("id malicioso não deveria ser aceito: %q", got)
		}
	})

	t.Run("substitui quando longo demais", func(t *testing.T) {
		w := do(r, "GET", "/", map[string]string{response.HeaderRequestID: strings.Repeat("a", 65)})
		if got := w.Header().Get(response.HeaderRequestID); len(got) != 36 {
			t.Errorf("id longo deveria ser substituído: %q", got)
		}
	})
}

// ---- Recovery --------------------------------------------------------------

func TestRecovery_ReturnsStandardErrorAndCountsPanic(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	m := telemetry.NewMetrics(nil)

	r := gin.New()
	r.Use(RequestID(), Recovery(log, m))
	r.GET("/boom", func(*gin.Context) { panic("falha inesperada") })

	w := do(r, "GET", "/boom", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", w.Code)
	}
	var body response.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("corpo não é JSON: %s", w.Body.String())
	}
	if body.Error.Code != "INTERNAL_ERROR" || body.Error.RequestID == "" {
		t.Errorf("corpo inesperado: %+v", body)
	}
	if strings.Contains(w.Body.String(), "falha inesperada") {
		t.Error("mensagem do panic não pode vazar para o cliente")
	}
	if !strings.Contains(logs.String(), "panic recuperado") || !strings.Contains(logs.String(), "stack") {
		t.Errorf("panic deveria ser logado com stack: %s", logs.String())
	}
	if got := testutil.ToFloat64(m.Panics); got != 1 {
		t.Errorf("métrica de panics = %v", got)
	}
}

// ---- Metrics ---------------------------------------------------------------

func TestMetrics_UsesRouteTemplateAsLabel(t *testing.T) {
	m := telemetry.NewMetrics(nil)
	r := gin.New()
	r.Use(Metrics(m))
	r.GET("/users/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	do(r, "GET", "/users/1", nil)
	do(r, "GET", "/users/2", nil)
	do(r, "GET", "/nao-existe", nil)

	if got := testutil.ToFloat64(m.RequestsTotal.WithLabelValues("GET", "/users/:id", "204")); got != 2 {
		t.Errorf("contador da rota = %v, esperado 2", got)
	}
	if got := testutil.ToFloat64(m.RequestsTotal.WithLabelValues("GET", "unmatched", "404")); got != 1 {
		t.Errorf("rotas inexistentes devem cair em 'unmatched', contador = %v", got)
	}
	if got := testutil.ToFloat64(m.InFlight); got != 0 {
		t.Errorf("in-flight deveria voltar a 0, obteve %v", got)
	}
}

// ---- AccessLog -------------------------------------------------------------

func TestAccessLog_LevelsAndFields(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	r := gin.New()
	r.Use(AccessLog(log, "/health"))
	r.GET("/health", func(c *gin.Context) { c.Status(200) })
	r.GET("/fail", func(c *gin.Context) { c.Status(503) })

	do(r, "GET", "/health", nil)
	if buf.Len() != 0 {
		t.Errorf("probe com sucesso deveria ser debug (silenciado): %s", buf.String())
	}

	do(r, "GET", "/fail", nil)
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log inválido: %s", buf.String())
	}
	if entry["level"] != "ERROR" || entry["route"] != "/fail" || entry["status"] != float64(503) {
		t.Errorf("entrada inesperada: %v", entry)
	}
}

// ---- RateLimit -------------------------------------------------------------

func TestRateLimit_BlocksAfterBurstPerIP(t *testing.T) {
	limiter := NewIPRateLimiter(1, 2)
	limited := 0
	r := gin.New()
	r.Use(RateLimit(limiter, func() { limited++ }))
	r.GET("/", func(c *gin.Context) { c.Status(200) })

	from := func(ip string) int {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	if from("10.0.0.1") != 200 || from("10.0.0.1") != 200 {
		t.Fatal("as 2 primeiras requisições (burst) deveriam passar")
	}
	if code := from("10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("3ª requisição deveria ser limitada, obteve %d", code)
	}
	if from("10.0.0.2") != 200 {
		t.Error("outro IP não pode ser afetado")
	}
	if limited != 1 {
		t.Errorf("callback de limite chamado %d vezes", limited)
	}
}

func TestIPRateLimiter_CleanupRemovesIdleVisitors(t *testing.T) {
	l := NewIPRateLimiter(1, 1)
	now := time.Now()
	l.now = func() time.Time { return now }

	l.Allow("a")
	l.Allow("b")
	now = now.Add(2 * time.Minute)
	l.Allow("b") // b continua ativo
	now = now.Add(2 * time.Minute)

	l.cleanup()

	if l.size() != 1 {
		t.Errorf("esperava apenas 'b' após limpeza, restaram %d", l.size())
	}
}

// ---- Segurança -------------------------------------------------------------

func TestSecurityHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders("/docs"))
	r.GET("/api", func(c *gin.Context) { c.Status(200) })
	r.GET("/docs/index.html", func(c *gin.Context) { c.Status(200) })

	w := do(r, "GET", "/api", nil)
	if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("headers ausentes: %v", w.Header())
	}
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Error("CSP deveria ser aplicada à API")
	}
	if do(r, "GET", "/docs/index.html", nil).Header().Get("Content-Security-Policy") != "" {
		t.Error("CSP restritiva quebraria a UI do Swagger")
	}
}

func TestCORS(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"https://app.workevent.com"}))
	r.GET("/", func(c *gin.Context) { c.Status(200) })

	preflight := func(origin string) *httptest.ResponseRecorder {
		return do(r, "OPTIONS", "/", map[string]string{
			"Origin":                        origin,
			"Access-Control-Request-Method": "GET",
		})
	}

	ok := preflight("https://app.workevent.com")
	if ok.Header().Get("Access-Control-Allow-Origin") != "https://app.workevent.com" {
		t.Errorf("origem permitida deveria ser liberada: %v", ok.Header())
	}
	if ok.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("credenciais deveriam ser permitidas")
	}

	denied := preflight("https://malicioso.com")
	if denied.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("origem não listada não pode ser liberada")
	}
}
