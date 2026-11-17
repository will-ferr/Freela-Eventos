// Package postgres contém a conexão com o PostgreSQL e as implementações dos
// repositórios do domínio usando GORM.
package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/will-ferr/Freela-Eventos/backend/config"
)

// Open abre o pool de conexões, aplica limites, registra o plugin de tracing
// (cada query vira um span) e valida a conexão com um ping.
func Open(ctx context.Context, cfg config.DBConfig, log *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.URL()), &gorm.Config{
		Logger:                 newGormLogger(log, cfg.SlowQuery),
		SkipDefaultTransaction: true, // transações explícitas nos casos de uso
		NowFunc:                func() time.Time { return time.Now().UTC() },
		TranslateError:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("abrindo conexão com postgres: %w", err)
	}

	if err := db.Use(newTracingPlugin()); err != nil {
		return nil, fmt.Errorf("registrando tracing do GORM: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping no postgres: %w", err)
	}
	return db, nil
}

// HealthChecker implementa health.Checker para o PostgreSQL.
type HealthChecker struct{ db *gorm.DB }

func NewHealthChecker(db *gorm.DB) *HealthChecker { return &HealthChecker{db: db} }

func (h *HealthChecker) Name() string { return "postgres" }

func (h *HealthChecker) Check(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
