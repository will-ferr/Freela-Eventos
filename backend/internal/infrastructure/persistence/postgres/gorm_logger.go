package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormLogger direciona os logs do GORM para o slog, mantendo request_id e
// trace_id. Registra apenas erros e queries lentas: logar toda query em
// produção gera ruído e pode expor dados pessoais.
type gormLogger struct {
	log   *slog.Logger
	slow  time.Duration
	level gormlogger.LogLevel
}

func newGormLogger(log *slog.Logger, slow time.Duration) gormlogger.Interface {
	return &gormLogger{log: log.With(slog.String("component", "gorm")), slow: slow, level: gormlogger.Warn}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cp := *l
	cp.level = level
	return &cp
}

func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		l.log.InfoContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		l.log.WarnContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		l.log.ErrorContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error:
		sql, rows := fc()
		l.log.ErrorContext(ctx, "query falhou",
			slog.String("sql", sql), slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed), slog.Any("error", err))
	case l.slow > 0 && elapsed > l.slow && l.level >= gormlogger.Warn:
		sql, rows := fc()
		l.log.WarnContext(ctx, "query lenta",
			slog.String("sql", sql), slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed), slog.Duration("threshold", l.slow))
	case l.level >= gormlogger.Info:
		sql, rows := fc()
		l.log.DebugContext(ctx, "query",
			slog.String("sql", sql), slog.Int64("rows", rows), slog.Duration("elapsed", elapsed))
	}
}
