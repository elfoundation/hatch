package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	h.Debug = os.Getenv("HATCH_DEBUG") != ""
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("hatch starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("hatch server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Create a deadline for shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
