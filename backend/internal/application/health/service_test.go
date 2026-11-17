package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeChecker struct {
	name  string
	err   error
	delay time.Duration
}

func (f fakeChecker) Name() string { return f.name }

func (f fakeChecker) Check(ctx context.Context) error {
	select {
	case <-time.After(f.delay):
		return f.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestReadiness_AllUp(t *testing.T) {
	svc := NewService(BuildInfo{Version: "1.0.0"}, time.Second, nil,
		fakeChecker{name: "postgres"},
		fakeChecker{name: "cache"},
	)

	r := svc.Readiness(context.Background())

	if r.Status != StatusUp {
		t.Fatalf("status = %s, esperado up", r.Status)
	}
	if len(r.Checks) != 2 {
		t.Fatalf("esperava 2 checks, obteve %d", len(r.Checks))
	}
	if r.Build.Version != "1.0.0" {
		t.Errorf("build info não propagada: %+v", r.Build)
	}
}

func TestReadiness_OneDownMakesOverallDown(t *testing.T) {
	svc := NewService(BuildInfo{}, time.Second, nil,
		fakeChecker{name: "postgres", err: errors.New("dial tcp: connection refused")},
		fakeChecker{name: "cache"},
	)

	r := svc.Readiness(context.Background())

	if r.Status != StatusDown {
		t.Fatalf("status = %s, esperado down", r.Status)
	}
	pg := r.Checks["postgres"]
	if pg.Status != StatusDown {
		t.Errorf("postgres deveria estar down")
	}
	if pg.Error != "indisponível" {
		t.Errorf("erro interno não pode vazar na resposta: %q", pg.Error)
	}
	if r.Checks["cache"].Status != StatusUp {
		t.Errorf("cache deveria continuar up")
	}
}

func TestReadiness_TimeoutIsReported(t *testing.T) {
	svc := NewService(BuildInfo{}, 20*time.Millisecond, nil,
		fakeChecker{name: "slow", delay: time.Second},
	)

	start := time.Now()
	r := svc.Readiness(context.Background())

	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("readiness deveria respeitar o timeout, levou %v", elapsed)
	}
	if r.Checks["slow"].Error != "tempo de resposta excedido" {
		t.Errorf("erro = %q", r.Checks["slow"].Error)
	}
}

func TestReadiness_NoCheckersIsUp(t *testing.T) {
	r := NewService(BuildInfo{}, time.Second, nil).Readiness(context.Background())
	if r.Status != StatusUp {
		t.Errorf("sem dependências o serviço deve estar up")
	}
}
