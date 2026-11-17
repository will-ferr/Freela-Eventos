package config

import (
	"strings"
	"testing"
	"time"
)

func baseEnv() map[string]string {
	return map[string]string{"DB_PASSWORD": "secret"}
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(LookupFromMap(baseEnv()))
	if err != nil {
		t.Fatalf("esperava sucesso, obteve %v", err)
	}
	if cfg.Env != EnvDevelopment {
		t.Errorf("Env = %q, esperado %q", cfg.Env, EnvDevelopment)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, esperado 8080", cfg.HTTP.Port)
	}
	if cfg.DB.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("ConnMaxLifetime = %v", cfg.DB.ConnMaxLifetime)
	}
	if !cfg.RateLimit.Enabled {
		t.Error("rate limit deveria vir habilitado por padrão")
	}
}

func TestLoad_ParsesValues(t *testing.T) {
	env := baseEnv()
	env["APP_ENV"] = "test"
	env["HTTP_PORT"] = "9090"
	env["CORS_ALLOWED_ORIGINS"] = " https://a.com , ,https://b.com "
	env["HTTP_READ_TIMEOUT"] = "3s"
	env["OTEL_TRACING_ENABLED"] = "true"
	env["OTEL_SAMPLE_RATIO"] = "0.25"

	cfg, err := Load(LookupFromMap(env))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.HTTP.Port != 9090 || cfg.HTTP.ReadTimeout != 3*time.Second {
		t.Errorf("HTTP não carregado corretamente: %+v", cfg.HTTP)
	}
	if got := cfg.CORS.AllowedOrigins; len(got) != 2 || got[0] != "https://a.com" || got[1] != "https://b.com" {
		t.Errorf("origens CORS = %v", got)
	}
	if !cfg.Telemetry.TracingEnabled || cfg.Telemetry.SampleRatio != 0.25 {
		t.Errorf("telemetria = %+v", cfg.Telemetry)
	}
}

func TestLoad_ReportsAllErrorsAtOnce(t *testing.T) {
	env := map[string]string{
		"HTTP_PORT":         "abc",
		"LOG_LEVEL":         "verbose",
		"HTTP_READ_TIMEOUT": "dez segundos",
	}
	_, err := Load(LookupFromMap(env))
	if err == nil {
		t.Fatal("esperava erro de validação")
	}
	for _, want := range []string{"HTTP_PORT", "LOG_LEVEL", "HTTP_READ_TIMEOUT", "DB_PASSWORD"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("erro deveria mencionar %s: %v", want, err)
		}
	}
}

func TestLoad_ProductionRules(t *testing.T) {
	env := baseEnv()
	env["APP_ENV"] = "production"
	env["CORS_ALLOWED_ORIGINS"] = "*"
	env["DB_SSLMODE"] = "disable"

	_, err := Load(LookupFromMap(env))
	if err == nil {
		t.Fatal("produção com CORS '*' e sslmode=disable deveria falhar")
	}
	if !strings.Contains(err.Error(), "CORS") || !strings.Contains(err.Error(), "DB_SSLMODE") {
		t.Errorf("mensagem incompleta: %v", err)
	}
}

func TestDBConfig_URLEscapesCredentials(t *testing.T) {
	db := DBConfig{Host: "db", Port: 5432, User: "user", Password: "p@ss:w/rd", Name: "app", SSLMode: "disable"}
	got := db.URL()
	want := "postgres://user:p%40ss%3Aw%2Frd@db:5432/app?sslmode=disable"
	if got != want {
		t.Errorf("URL()\n obtido:   %s\n esperado: %s", got, want)
	}
}
