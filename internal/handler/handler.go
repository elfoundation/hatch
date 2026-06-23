package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/elfoundation/hatch/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler groups all HTTP handlers and their shared dependencies.
type Handler struct {
	Repo store.Repository
}

// New creates a new Handler with the given store.
func New(repo store.Repository) *Handler {
	return &Handler{Repo: repo}
}

// RegisterRoutes mounts all routes on the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check.
	r.Get("/healthz", Healthz)

	// JSON API v1 routes.
	h.RegisterV1Routes(r)

	// Inspect page: server-rendered request list.
	r.Get("/e/{endpointID}", HandleInspect(h.Repo))

	// SSE stream for live updates.
	r.Get("/e/{endpointID}/events", HandleSSE(h.Repo))

	// Mock configuration.
	r.Put("/e/{endpointID}/mock", HandleMock(h.Repo))

	// Replay: one-click re-send of a captured request.
	r.Post("/e/{endpointID}/requests/{requestID}/replay", HandleReplay(h.Repo))

	// Capture webhook: any method on /{endpointID} (wildcard, must be last).
	r.HandleFunc("/{endpointID}", h.capture)
	r.HandleFunc("/{endpointID}/*", h.capture)
}

// capture handles incoming webhook captures on /{endpointID}.
func (h *Handler) capture(w http.ResponseWriter, r *http.Request) {
	eid := r.PathValue("endpointID")
	if eid == "" {
		writeError(w, http.StatusBadRequest, "missing endpoint ID")
		return
	}

	ctx := r.Context()

	// Auto-create endpoint if it doesn't exist.
	if _, err := h.Repo.GetEndpoint(ctx, eid); err != nil {
		h.Repo.CreateEndpoint(ctx, eid)
	}

	// Collect headers as JSON.
	hdr := map[string]string{}
	for k, v := range r.Header {
		hdr[k] = strings.Join(v, ", ")
	}
	hdrJSON, _ := json.Marshal(hdr)

	// Read body.
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	// Store the request.
	req := &store.Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: string(hdrJSON),
		Query:   r.URL.RawQuery,
		Body:    body,
	}
	if err := h.Repo.AppendRequest(ctx, eid, req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store request")
		return
	}

	// Broadcast via SSE.
	broadcastRequest(eid, req)

	// Look up mock response for this endpoint.
	mock, err := h.Repo.GetMock(ctx, eid)
	if err == nil && mock != nil {
		for k, v := range mock.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(mock.Status)
		if mock.Body != nil {
			w.Write(mock.Body)
		}
		return
	}

	// Default: return 200 OK with empty JSON.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// Healthz returns 200 OK for liveness probes.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
