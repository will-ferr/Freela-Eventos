// Package health implementa o caso de uso de verificação de saúde da aplicação.
// Ele não sabe o que está verificando: cada dependência (Postgres, futuramente
// cache, fila...) entra como um Checker.
package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

// Checker verifica uma dependência externa.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

type CheckResult struct {
	Status    Status `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

type Report struct {
	Status    Status                 `json:"status"`
	Checks    map[string]CheckResult `json:"checks"`
	Build     BuildInfo              `json:"build"`
	Uptime    string                 `json:"uptime"`
	CheckedAt time.Time              `json:"checked_at"`
}

type Service struct {
	checkers []Checker
	timeout  time.Duration
	build    BuildInfo
	started  time.Time
	now      func() time.Time
	logger   *slog.Logger
}

func NewService(build BuildInfo, timeout time.Duration, logger *slog.Logger, checkers ...Checker) *Service {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		checkers: checkers,
		timeout:  timeout,
		build:    build,
		started:  time.Now(),
		now:      time.Now,
		logger:   logger,
	}
}

// Readiness executa todas as verificações em paralelo, cada uma limitada pelo
// timeout configurado. O status geral só é "up" se todas estiverem "up".
func (s *Service) Readiness(ctx context.Context) Report {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	results := make(map[string]CheckResult, len(s.checkers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, c := range s.checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			start := time.Now()
			err := c.Check(ctx)
			res := CheckResult{Status: StatusUp, LatencyMs: time.Since(start).Milliseconds()}
			if err != nil {
				res.Status = StatusDown
				// Mensagem genérica: detalhes de infraestrutura vão para o log,
				// nunca para a resposta pública.
				res.Error = "indisponível"
				if ctx.Err() != nil {
					res.Error = "tempo de resposta excedido"
				}
				s.logger.WarnContext(ctx, "health check falhou",
					slog.String("check", c.Name()), slog.Any("error", err))
			}
			mu.Lock()
			results[c.Name()] = res
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	overall := StatusUp
	for _, r := range results {
		if r.Status != StatusUp {
			overall = StatusDown
			break
		}
	}

	now := s.now()
	return Report{
		Status:    overall,
		Checks:    results,
		Build:     s.build,
		Uptime:    now.Sub(s.started).Round(time.Second).String(),
		CheckedAt: now.UTC(),
	}
}
