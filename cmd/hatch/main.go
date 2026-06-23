package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/elfoundation/hatch/internal/handler"
	"github.com/elfoundation/hatch/internal/store"
	"github.com/go-chi/chi/v5"
)

func main() {
	// CLI mode: if a subcommand is provided, handle it and exit.
	if cliMain() {
		return
	}

	// Server mode: start the HTTP server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("HATCH_DB_PATH")
	repo, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("hatch: open store: %v", err)
	}
	defer repo.Close()

	h := handler.New(repo)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hatch starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("hatch server error: %v", err)
	}
}
