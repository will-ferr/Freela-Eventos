// Command api inicia o servidor HTTP do WorkEvent.
//
//	@title						WorkEvent API
//	@version					1.0
//	@description				API da plataforma de gestão de freelancers para eventos.
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/will-ferr/Freela-Eventos/backend/config"
	"github.com/will-ferr/Freela-Eventos/backend/internal/application/health"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/logger"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/persistence/postgres"
	"github.com/will-ferr/Freela-Eventos/backend/internal/infrastructure/telemetry"
	httpapi "github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http"
	"github.com/will-ferr/Freela-Eventos/backend/internal/interfaces/http/middleware"
)

// Preenchidas via -ldflags no build (ver Dockerfile).
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "verifica /health/live e encerra (usado pelo Docker)")
	flag.Parse()

	if *healthcheck {
		os.Exit(runHealthcheck())
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	log := logger.New(os.Stdout, cfg.Log.Level, cfg.Log.Format, cfg.Telemetry.ServiceName, cfg.Env)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.SetupTracing(ctx, telemetry.TracingOptions{
		Enabled:     cfg.Telemetry.TracingEnabled,
		ServiceName: cfg.Telemetry.ServiceName,
		Version:     version,
		Env:         cfg.Env,
		Endpoint:    cfg.Telemetry.OTLPEndpoint,
		SampleRatio: cfg.Telemetry.SampleRatio,
	})
	if err != nil {
		return err
	}

	db, err := postgres.Open(ctx, cfg.DB, log)
	if err != nil {
		return err
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	build := health.BuildInfo{Version: version, Commit: commit, BuildTime: buildTime}
	metrics := telemetry.NewMetrics(map[string]string{"version": version, "commit": commit})
	metrics.RegisterDB(sqlDB, cfg.DB.Name)

	limiter := middleware.NewIPRateLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst)
	go limiter.Run(ctx)

	router, err := httpapi.NewRouter(httpapi.Deps{
		Config:      cfg,
		Logger:      log,
		Metrics:     metrics,
		Health:      health.NewService(build, 2*time.Second, log, postgres.NewHealthChecker(db)),
		RateLimiter: limiter,
	})
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           router,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("servidor iniciado",
			slog.Int("port", cfg.HTTP.Port), slog.String("version", version), slog.String("commit", commit),
			slog.Bool("tracing", cfg.Telemetry.TracingEnabled))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// Graceful shutdown: para de aceitar conexões, espera as requisições em
	// andamento e só então descarrega os spans pendentes.
	log.Info("encerrando servidor")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("falha no shutdown do HTTP", slog.Any("error", err))
	}
	if err := shutdownTracing(shutdownCtx); err != nil {
		log.Error("falha ao descarregar traces", slog.Any("error", err))
	}
	log.Info("servidor encerrado")
	return nil
}

// runHealthcheck existe porque a imagem final (distroless) não tem curl/wget.
func runHealthcheck() int {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + httpapi.APIPrefix + "/health/live")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
