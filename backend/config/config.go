// Package config carrega e valida a configuração da aplicação a partir de
// variáveis de ambiente. Nenhum outro pacote deve ler os.Getenv diretamente.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type Config struct {
	Env       string
	HTTP      HTTPConfig
	DB        DBConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Log       LogConfig
	Telemetry TelemetryConfig
	Swagger   bool
}

type HTTPConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SlowQuery       time.Duration
}

// URL retorna a connection string no formato URL, aceita tanto pelo pgx
// quanto pelo golang-migrate. Usuário e senha são escapados corretamente.
func (d DBConfig) URL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.Name,
	}
	q := u.Query()
	q.Set("sslmode", d.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	Enabled bool
	RPS     float64
	Burst   int
}

type LogConfig struct {
	Level  string
	Format string // json | text
}

type TelemetryConfig struct {
	ServiceName    string
	TracingEnabled bool
	OTLPEndpoint   string // host:porta do coletor OTLP/HTTP
	SampleRatio    float64
}

// Load lê as variáveis de ambiente usando a função lookup informada
// (os.LookupEnv em produção; um map nos testes) e valida o resultado.
func Load(lookup func(string) (string, bool)) (*Config, error) {
	r := reader{lookup: lookup}

	cfg := &Config{
		Env: r.str("APP_ENV", EnvDevelopment),
		HTTP: HTTPConfig{
			Port:            r.int("HTTP_PORT", 8080),
			ReadTimeout:     r.duration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    r.duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     r.duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: r.duration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			TrustedProxies:  r.list("HTTP_TRUSTED_PROXIES", nil),
		},
		DB: DBConfig{
			Host:            r.str("DB_HOST", "localhost"),
			Port:            r.int("DB_PORT", 5432),
			User:            r.str("DB_USER", "workevent"),
			Password:        r.str("DB_PASSWORD", ""),
			Name:            r.str("DB_NAME", "workevent"),
			SSLMode:         r.str("DB_SSLMODE", "disable"),
			MaxOpenConns:    r.int("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    r.int("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: r.duration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			SlowQuery:       r.duration("DB_SLOW_QUERY_THRESHOLD", 200*time.Millisecond),
		},
		CORS: CORSConfig{
			AllowedOrigins: r.list("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		RateLimit: RateLimitConfig{
			Enabled: r.bool("RATE_LIMIT_ENABLED", true),
			RPS:     r.float("RATE_LIMIT_RPS", 10),
			Burst:   r.int("RATE_LIMIT_BURST", 30),
		},
		Log: LogConfig{
			Level:  r.str("LOG_LEVEL", "info"),
			Format: r.str("LOG_FORMAT", "json"),
		},
		Telemetry: TelemetryConfig{
			ServiceName:    r.str("OTEL_SERVICE_NAME", "workevent-api"),
			TracingEnabled: r.bool("OTEL_TRACING_ENABLED", false),
			OTLPEndpoint:   r.str("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
			SampleRatio:    r.float("OTEL_SAMPLE_RATIO", 1.0),
		},
		Swagger: r.bool("SWAGGER_ENABLED", true),
	}

	if err := errors.Join(append(r.errs, cfg.validate()...)...); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

func (c *Config) validate() []error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	switch c.Env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		add("APP_ENV deve ser development, test ou production (recebido %q)", c.Env)
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		add("HTTP_PORT fora do intervalo válido: %d", c.HTTP.Port)
	}
	if c.DB.Password == "" {
		add("DB_PASSWORD é obrigatório")
	}
	if c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		add("DB_MAX_IDLE_CONNS (%d) não pode ser maior que DB_MAX_OPEN_CONNS (%d)", c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		add("LOG_LEVEL inválido: %q", c.Log.Level)
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		add("LOG_FORMAT deve ser json ou text (recebido %q)", c.Log.Format)
	}
	if c.RateLimit.Enabled && (c.RateLimit.RPS <= 0 || c.RateLimit.Burst <= 0) {
		add("RATE_LIMIT_RPS e RATE_LIMIT_BURST devem ser positivos")
	}
	if c.Telemetry.SampleRatio < 0 || c.Telemetry.SampleRatio > 1 {
		add("OTEL_SAMPLE_RATIO deve estar entre 0 e 1")
	}
	if c.IsProduction() {
		for _, o := range c.CORS.AllowedOrigins {
			if o == "*" {
				add("CORS_ALLOWED_ORIGINS não pode conter '*' em produção")
			}
		}
		if c.DB.SSLMode == "disable" {
			add("DB_SSLMODE=disable não é permitido em produção")
		}
	}
	return errs
}

// LookupFromMap facilita testes e ferramentas que montam o ambiente manualmente.
func LookupFromMap(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

// normaliza listas separadas por vírgula, ignorando itens vazios.
func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
