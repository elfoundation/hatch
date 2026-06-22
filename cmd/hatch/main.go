// Command hatch is the Hatch HTTP request inspector + mocker server.
//
// Hatch captures, inspects, and mocks HTTP requests. It ships as a single
// static binary — one command on a VPS and your payloads never leave your
// network.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("hatch starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("hatch server error: %v", err)
	}
}

// healthz returns 200 OK with body "ok" for liveness probes.
func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}
