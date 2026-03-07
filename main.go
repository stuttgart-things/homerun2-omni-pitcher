package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/config"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/middleware"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library"
)

// Build-time variables set via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	config.SetupLogging()

	port := homerun.GetEnv("PORT", "8080")
	mode := homerun.GetEnv("PITCHER_MODE", "redis")

	// Select pitcher backend
	var p pitcher.Pitcher
	switch mode {
	case "file":
		filePath := homerun.GetEnv("PITCHER_FILE", "pitched.log")
		p = &pitcher.FilePitcher{Path: filePath}
		slog.Info("pitcher mode: file", "path", filePath)
	default:
		redisConfig := config.LoadRedisConfig()
		rp := &pitcher.RedisPitcher{Config: redisConfig}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := rp.HealthCheck(ctx); err != nil {
			slog.Error("redis health check failed", "error", err, "addr", redisConfig.Addr, "port", redisConfig.Port)
			cancel()
			os.Exit(1)
		}
		cancel()

		p = rp
		slog.Info("pitcher mode: redis", "addr", redisConfig.Addr, "port", redisConfig.Port, "stream", redisConfig.Stream)
	}

	buildInfo := handlers.BuildInfo{Version: version, Commit: commit, Date: date}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.NewHealthHandler(buildInfo))
	mux.HandleFunc("/pitch", middleware.TokenAuthMiddleware(handlers.NewPitchHandler(p)))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting server",
			"app", "homerun2-omni-pitcher",
			"version", version,
			"commit", commit,
			"date", date,
			"port", port,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}
