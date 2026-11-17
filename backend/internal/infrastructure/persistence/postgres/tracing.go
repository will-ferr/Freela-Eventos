package postgres

import (
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// tracingPlugin cria um span por operação do GORM. É uma implementação
// própria e enxuta: o plugin oficial importa drivers de MySQL e ClickHouse
// que não usamos, inflando o binário.
type tracingPlugin struct{ tracer trace.Tracer }

func newTracingPlugin() gorm.Plugin {
	return &tracingPlugin{tracer: otel.Tracer("github.com/will-ferr/Freela-Eventos/backend/postgres")}
}

func (p *tracingPlugin) Name() string { return "workevent:tracing" }

const spanKey = "workevent:span"

func (p *tracingPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	hooks := []struct {
		op     string
		before func(string, func(*gorm.DB)) error
		after  func(string, func(*gorm.DB)) error
	}{
		{"create", cb.Create().Before("gorm:create").Register, cb.Create().After("gorm:create").Register},
		{"query", cb.Query().Before("gorm:query").Register, cb.Query().After("gorm:query").Register},
		{"update", cb.Update().Before("gorm:update").Register, cb.Update().After("gorm:update").Register},
		{"delete", cb.Delete().Before("gorm:delete").Register, cb.Delete().After("gorm:delete").Register},
		{"row", cb.Row().Before("gorm:row").Register, cb.Row().After("gorm:row").Register},
		{"raw", cb.Raw().Before("gorm:raw").Register, cb.Raw().After("gorm:raw").Register},
	}
	for _, h := range hooks {
		if err := h.before("otel:before_"+h.op, p.start(h.op)); err != nil {
			return err
		}
		if err := h.after("otel:after_"+h.op, p.end); err != nil {
			return err
		}
	}
	return nil
}

func (p *tracingPlugin) start(op string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Context == nil {
			return
		}
		ctx, span := p.tracer.Start(db.Statement.Context, "db."+op,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "postgresql"),
				attribute.String("db.operation", op),
			),
		)
		db.Statement.Context = ctx
		db.InstanceSet(spanKey, span)
	}
}

func (p *tracingPlugin) end(db *gorm.DB) {
	v, ok := db.InstanceGet(spanKey)
	if !ok {
		return
	}
	span, ok := v.(trace.Span)
	if !ok {
		return
	}
	defer span.End()

	if db.Statement.Table != "" {
		span.SetAttributes(attribute.String("db.sql.table", db.Statement.Table))
	}
	// SQL com placeholders ($1, $2...): nunca inclui os valores (dados pessoais).
	span.SetAttributes(
		attribute.String("db.statement", db.Statement.SQL.String()),
		attribute.Int64("db.rows_affected", db.RowsAffected),
	)
	if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
		span.RecordError(db.Error)
		span.SetStatus(codes.Error, db.Error.Error())
	}
}
