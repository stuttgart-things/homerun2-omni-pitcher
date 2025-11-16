package main

import (
	"log"
	"net/http"

	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/handlers"
	"github.com/stuttgart-things/homerun2-omni-pitcher/internal/middleware"

	homerun "github.com/stuttgart-things/homerun-library"
)

func main() {
	port := homerun.GetEnv("PORT", "8080")

	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/pitch", middleware.TokenAuthMiddleware(handlers.PitchHandler))

	log.Printf("Starting homerun2-omni-pitcher on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
