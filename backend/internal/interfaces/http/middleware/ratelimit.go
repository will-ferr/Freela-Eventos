package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/response"
)

// IPRateLimiter aplica token bucket por chave (IP do cliente).
//
// Limitação conhecida (ADR-004): o estado fica em memória, então cada réplica
// tem seu próprio limite. Com mais de uma instância, migrar para Redis.
type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
	ttl      time.Duration
	now      func() time.Time
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(rps),
		burst:    burst,
		ttl:      3 * time.Minute,
		now:      time.Now,
	}
}

func (l *IPRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	v, ok := l.visitors[key]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.visitors[key] = v
	}
	v.lastSeen = now
	return v.limiter.AllowN(now, 1)
}

// cleanup remove visitantes inativos para o mapa não crescer indefinidamente.
func (l *IPRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := l.now().Add(-l.ttl)
	for k, v := range l.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(l.visitors, k)
		}
	}
}

// Run executa a limpeza periódica até o contexto ser cancelado.
func (l *IPRateLimiter) Run(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.cleanup()
		}
	}
}

func (l *IPRateLimiter) size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.visitors)
}

// RateLimit rejeita com 429 quando o cliente excede o limite.
func RateLimit(l *IPRateLimiter, onLimited func()) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(c.ClientIP()) {
			if onLimited != nil {
				onLimited()
			}
			c.Header("Retry-After", "1")
			response.AbortWith(c, http.StatusTooManyRequests, "RATE_LIMITED",
				"Muitas requisições. Aguarde um instante e tente novamente.")
			return
		}
		c.Next()
	}
}
