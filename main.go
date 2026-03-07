package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/config"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/middleware"

	homerun "github.com/stuttgart-things/homerun-library"
)

// Build-time variables set via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	port := homerun.GetEnv("PORT", "8080")

	// Load config once at startup
	redisConfig := config.LoadRedisConfig()
	buildInfo := handlers.BuildInfo{Version: version, Commit: commit, Date: date}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.NewHealthHandler(buildInfo))
	mux.HandleFunc("/pitch", middleware.TokenAuthMiddleware(handlers.NewPitchHandler(redisConfig)))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting homerun2-omni-pitcher %s (%s, %s) on port %s", version, commit, date, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
