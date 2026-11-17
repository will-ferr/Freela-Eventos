package postgres

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func newTestLogger(buf *bytes.Buffer, slow time.Duration) *gormLogger {
	l := newGormLogger(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})), slow)
	return l.(*gormLogger)
}

func sqlFn() (string, int64) { return "SELECT 1", 1 }

func TestGormLogger_LogsSlowQueries(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, 10*time.Millisecond)

	l.Trace(context.Background(), time.Now().Add(-50*time.Millisecond), sqlFn, nil)

	if !strings.Contains(buf.String(), "query lenta") {
		t.Errorf("query lenta deveria ser registrada: %s", buf.String())
	}
}

func TestGormLogger_IgnoresFastQueriesAtWarn(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, time.Second)

	l.Trace(context.Background(), time.Now(), sqlFn, nil)

	if buf.Len() != 0 {
		t.Errorf("query rápida não deveria gerar log: %s", buf.String())
	}
}

func TestGormLogger_RecordNotFoundIsNotAnError(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, time.Second)

	l.Trace(context.Background(), time.Now(), sqlFn, gorm.ErrRecordNotFound)

	if buf.Len() != 0 {
		t.Errorf("ErrRecordNotFound é fluxo normal, não erro: %s", buf.String())
	}
}

func TestGormLogger_LogsErrors(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, time.Second)

	l.Trace(context.Background(), time.Now(), sqlFn, errors.New("duplicate key"))

	if !strings.Contains(buf.String(), "query falhou") || !strings.Contains(buf.String(), "duplicate key") {
		t.Errorf("erro de query deveria ser registrado: %s", buf.String())
	}
}
