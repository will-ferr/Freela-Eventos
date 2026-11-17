// Package middleware contém os middlewares HTTP transversais: correlação,
// logs de acesso, métricas, recuperação de panics, CORS, rate limit e headers
// de segurança.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/logger"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

// RequestID reaproveita o X-Request-ID recebido (ex.: gerado pelo Nginx) se
// for seguro, ou gera um novo. O id volta no header da resposta e é anexado ao
// contexto para aparecer em todos os logs da requisição.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(response.HeaderRequestID)
		if !validRequestID(id) {
			id = uuid.NewString()
		}
		c.Writer.Header().Set(response.HeaderRequestID, id)
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}

// validRequestID evita log injection e valores gigantes vindos do cliente.
func validRequestID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}
