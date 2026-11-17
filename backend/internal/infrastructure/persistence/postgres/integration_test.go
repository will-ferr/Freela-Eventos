//go:build integration

// Testes de integração: exigem um PostgreSQL real.
//
//	TEST_DATABASE_URL=postgres://workevent:workevent@localhost:5433/workevent_test?sslmode=disable \
//	  go test -tags=integration ./...
//
// Cada execução recria o schema do zero aplicando as migrations reais.
package postgres_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"gorm.io/gorm"

	"github.com/will-ferr/Freela-Eventos/backend/config"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/persistence/postgres"
	"github.com/will-ferr/Freela-Eventos/backend/migrations"
)

func dbConfig(t *testing.T) config.DBConfig {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL não definido")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL inválido: %v", err)
	}
	pass, _ := u.User.Password()
	port, _ := strconv.Atoi(u.Port())
	return config.DBConfig{
		Host: u.Hostname(), Port: port, User: u.User.Username(), Password: pass,
		Name: strings.TrimPrefix(u.Path, "/"), SSLMode: u.Query().Get("sslmode"),
		MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute, SlowQuery: time.Second,
	}
}

// setup derruba e recria o schema, garantindo que up e down funcionam.
func setup(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := dbConfig(t)

	m, err := migrations.New(cfg.URL())
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Drop(); err != nil {
		t.Fatalf("drop: %v", err)
	}
	m.Close()
	if err := migrations.Up(cfg.URL()); err != nil {
		t.Fatalf("migrations up: %v", err)
	}

	db, err := postgres.Open(context.Background(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() {
		if s, err := db.DB(); err == nil {
			s.Close()
		}
	})
	return db
}

func TestMigrations_DownAndUpAreReversible(t *testing.T) {
	cfg := dbConfig(t)
	setup(t)

	m, err := migrations.New(cfg.URL())
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		t.Fatalf("down falhou — migration .down.sql quebrada: %v", err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("up após down falhou: %v", err)
	}
	if _, dirty, _ := m.Version(); dirty {
		t.Fatal("banco ficou em estado dirty")
	}
}

func TestCategories_SeededAndConstrained(t *testing.T) {
	db := setup(t)
	ctx := context.Background()

	var count int64
	if err := db.WithContext(ctx).Table("professional_categories").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 16 {
		t.Errorf("esperava 16 categorias iniciais, obteve %d", count)
	}

	t.Run("slug duplicado é rejeitado", func(t *testing.T) {
		err := db.Exec(`INSERT INTO professional_categories (name, slug) VALUES ('Outro Garçom', 'garcom')`).Error
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			t.Errorf("esperava ErrDuplicatedKey, obteve %v", err)
		}
	})

	t.Run("nome duplicado ignora maiúsculas", func(t *testing.T) {
		err := db.Exec(`INSERT INTO professional_categories (name, slug) VALUES ('BARTENDER', 'bartender-2')`).Error
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			t.Errorf("esperava ErrDuplicatedKey, obteve %v", err)
		}
	})

	t.Run("slug com formato inválido é rejeitado", func(t *testing.T) {
		err := db.Exec(`INSERT INTO professional_categories (name, slug) VALUES ('Teste', 'Slug Inválido')`).Error
		if err == nil || !strings.Contains(err.Error(), "check") {
			t.Errorf("esperava violação de CHECK, obteve %v", err)
		}
	})

	t.Run("exclusão lógica libera o slug", func(t *testing.T) {
		if err := db.Exec(`UPDATE professional_categories SET deleted_at = now() WHERE slug = 'barback'`).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO professional_categories (name, slug) VALUES ('Barback', 'barback')`).Error; err != nil {
			t.Errorf("slug de registro excluído deveria poder ser reutilizado: %v", err)
		}
	})

	t.Run("trigger atualiza updated_at", func(t *testing.T) {
		var before, after time.Time
		db.Raw(`SELECT updated_at FROM professional_categories WHERE slug = 'caixa'`).Scan(&before)
		time.Sleep(10 * time.Millisecond)
		db.Exec(`UPDATE professional_categories SET description = 'novo' WHERE slug = 'caixa'`)
		db.Raw(`SELECT updated_at FROM professional_categories WHERE slug = 'caixa'`).Scan(&after)
		if !after.After(before) {
			t.Errorf("updated_at não mudou: antes=%v depois=%v", before, after)
		}
	})
}

func TestHealthChecker_Postgres(t *testing.T) {
	db := setup(t)
	hc := postgres.NewHealthChecker(db)

	if err := hc.Check(context.Background()); err != nil {
		t.Errorf("banco disponível deveria passar: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.Close()
	if err := hc.Check(context.Background()); err == nil {
		t.Error("conexão fechada deveria falhar no health check")
	}
}

func TestTracingPlugin_RecordsSpansWithoutValues(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	db := setup(t) // abre depois de trocar o provider para o plugin usar o recorder

	var name string
	db.WithContext(context.Background()).
		Table("professional_categories").
		Select("name").
		Where("slug = ?", "dj-entretenimento").
		Scan(&name)

	var found bool
	for _, s := range recorder.Ended() {
		if s.Name() != "db.row" && s.Name() != "db.query" {
			continue
		}
		for _, a := range s.Attributes() {
			if a.Key == "db.statement" {
				found = true
				if strings.Contains(a.Value.AsString(), "dj-entretenimento") {
					t.Errorf("valor do parâmetro vazou no span: %s", a.Value.AsString())
				}
			}
		}
	}
	if !found {
		t.Errorf("nenhum span de query registrado; spans: %d", len(recorder.Ended()))
	}
}
