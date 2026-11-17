package middleware

import (
	"slices"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

// SecurityHeaders aplica headers defensivos. A API só devolve JSON, então a
// CSP é a mais restritiva possível — exceto na UI do Swagger, que precisa
// carregar scripts e estilos próprios.
func SecurityHeaders(swaggerPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		if swaggerPrefix == "" || !strings.HasPrefix(c.Request.URL.Path, swaggerPrefix) {
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}
		c.Next()
	}
}

// CORS libera apenas as origens configuradas. Credenciais ficam habilitadas
// porque o refresh token trafegará em cookie httpOnly (ADR-003).
func CORS(origins []string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", response.HeaderRequestID, "traceparent"},
		ExposeHeaders:    []string{response.HeaderRequestID, "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if slices.Contains(origins, "*") {
		// Apenas para desenvolvimento; bloqueado em produção pela validação de config.
		cfg.AllowOriginFunc = func(string) bool { return true }
	} else {
		cfg.AllowOrigins = origins
	}
	return cors.New(cfg)
}
