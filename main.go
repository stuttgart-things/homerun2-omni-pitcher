package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/banner"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/config"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/middleware"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/pitcher"

	homerun "github.com/stuttgart-things/homerun-library/v2"
)

// Build-time variables set via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	banner.Show()
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

		// Ensure RediSearch index exists before accepting requests
		if err := rp.EnsureIndex(context.Background()); err != nil {
			slog.Warn("failed to ensure redisearch index", "index", redisConfig.Index, "error", err)
		}

		p = rp
		slog.Info("pitcher mode: redis", "addr", redisConfig.Addr, "port", redisConfig.Port, "stream", redisConfig.Stream, "searchIndex", redisConfig.Index)
	}

	// Select auth middleware
	authMode := homerun.GetEnv("AUTH_MODE", "token")
	var authMiddleware func(http.HandlerFunc) http.HandlerFunc

	switch authMode {
	case "jwt":
		jwksURL := homerun.GetEnv("JWT_JWKS_URL", "")
		if jwksURL == "" {
			slog.Error("JWT_JWKS_URL is required when AUTH_MODE=jwt")
			os.Exit(1)
		}
		jwtMw, err := middleware.NewJWTAuthMiddleware(middleware.JWTConfig{
			JWKSURL:  jwksURL,
			Issuer:   homerun.GetEnv("JWT_ISSUER", ""),
			Audience: homerun.GetEnv("JWT_AUDIENCE", ""),
		})
		if err != nil {
			slog.Error("failed to initialize JWT auth", "error", err)
			os.Exit(1)
		}
		authMiddleware = jwtMw
	default:
		authMiddleware = middleware.TokenAuthMiddleware
	}

	// Startup banner
	slog.Info("starting homerun2-omni-pitcher",
		"version", version,
		"commit", commit,
		"date", date,
		"go", runtime.Version(),
		"port", port,
		"pitcher_mode", mode,
		"auth_mode", authMode,
	)

	buildInfo := handlers.BuildInfo{Version: version, Commit: commit, Date: date}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.NewHealthHandler(buildInfo))
	mux.HandleFunc("/pitch", authMiddleware(handlers.NewPitchHandler(p)))
	mux.HandleFunc("/pitch/grafana", authMiddleware(handlers.NewGrafanaPitchHandler(p)))

	githubWebhookSecret := homerun.GetEnv("GITHUB_WEBHOOK_SECRET", "")
	mux.HandleFunc("/pitch/github", authMiddleware(handlers.NewGitHubPitchHandler(p, githubWebhookSecret)))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: middleware.RequestLogging(mux),
	}

	// Start server in goroutine
	go func() {
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
