// Package http monta o roteador Gin com middlewares e rotas.
package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/will-ferr/Freela-Eventos/backend/config"
	_ "github.com/will-ferr/Freela-Eventos/backend/docs" // spec Swagger gerada
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/telemetry"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/handler"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/middleware"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

const (
	APIPrefix  = "/api/v1"
	docsPrefix = APIPrefix + "/docs"
)

type Deps struct {
	Config      *config.Config
	Logger      *slog.Logger
	Metrics     *telemetry.Metrics
	Health      handler.ReadinessChecker
	RateLimiter *middleware.IPRateLimiter
}

func NewRouter(d Deps) (*gin.Engine, error) {
	if d.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true
	// Sem proxies confiáveis, ClientIP() usa o RemoteAddr; atrás do Nginx é
	// preciso configurar a rede dele, senão todos os usuários teriam o mesmo IP.
	if err := r.SetTrustedProxies(d.Config.HTTP.TrustedProxies); err != nil {
		return nil, err
	}

	liveRoute := APIPrefix + "/health/live"
	readyRoute := APIPrefix + "/health/ready"

	// Ordem importa: o span do OTel nasce primeiro (trace_id disponível para
	// os logs); Recovery fica dentro de AccessLog/Metrics para que um panic
	// ainda seja registrado como 500.
	r.Use(
		otelgin.Middleware(d.Config.Telemetry.ServiceName,
			otelgin.WithFilter(func(req *http.Request) bool {
				p := req.URL.Path
				return p != liveRoute && p != readyRoute && p != "/metrics"
			}),
		),
		middleware.RequestID(),
		middleware.AccessLog(d.Logger, liveRoute, readyRoute, "/metrics"),
		middleware.Metrics(d.Metrics),
		middleware.Recovery(d.Logger, d.Metrics),
		middleware.SecurityHeaders(docsPrefix),
		middleware.CORS(d.Config.CORS.AllowedOrigins),
	)

	// /metrics fica fora de /api: o Nginx só expõe /api publicamente, então o
	// endpoint é acessível apenas pela rede interna (Prometheus).
	r.GET("/metrics", gin.WrapH(d.Metrics.Handler()))

	api := r.Group(APIPrefix)
	if d.Config.RateLimit.Enabled && d.RateLimiter != nil {
		api.Use(middleware.RateLimit(d.RateLimiter, d.Metrics.RateLimited.Inc))
	}

	healthH := handler.NewHealthHandler(d.Health)
	api.GET("/health/live", healthH.Live)
	api.GET("/health/ready", healthH.Ready)

	if d.Config.Swagger {
		api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	r.NoRoute(func(c *gin.Context) {
		response.AbortWith(c, http.StatusNotFound, "ROUTE_NOT_FOUND", "Recurso não encontrado.")
	})
	r.NoMethod(func(c *gin.Context) {
		response.AbortWith(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método não permitido para este recurso.")
	})

	return r, nil
}
