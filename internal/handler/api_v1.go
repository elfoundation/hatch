package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/elfoundation/hatch/internal/store"
	"github.com/go-chi/chi/v5"
)

// RegisterV1Routes mounts the JSON API v1 routes on the given router.
func (h *Handler) RegisterV1Routes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Route("/endpoints/{endpointID}", func(r chi.Router) {
			// Capture request (JSON body)
			r.Post("/requests", h.v1CaptureRequest)
			// List or search requests
			r.Get("/requests", h.v1ListRequests)
			// Replay a specific request
			r.Post("/requests/{requestID}/replay", h.v1ReplayRequest)
			// Mock configuration
			r.Put("/mock", h.v1SetMock)
			// OpenAPI spec
			r.Get("/openapi.json", h.v1OpenAPI)
		})
	})
}

// v1CaptureRequest handles POST /v1/endpoints/{endpointID}/requests
func (h *Handler) v1CaptureRequest(w http.ResponseWriter, r *http.Request) {
	endpointID := chi.URLParam(r, "endpointID")
	if endpointID == "" {
		writeError(w, http.StatusBadRequest, "missing endpoint ID")
		return
	}
	ctx := r.Context()

	// Auto-create endpoint if it doesn't exist.
	if _, err := h.Repo.GetEndpoint(ctx, endpointID); err != nil {
		h.Repo.CreateEndpoint(ctx, endpointID)
	}

	// Parse JSON body representing a request.
	var incoming struct {
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Headers map[string]string `json:"headers"`
		Query   string            `json:"query"`
		Body    string            `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if incoming.Method == "" {
		incoming.Method = "POST" // default
	}
	if incoming.Path == "" {
		incoming.Path = "/"
	}

	// Convert headers map to JSON string.
	headersJSON := "{}"
	if incoming.Headers != nil {
		b, _ := json.Marshal(incoming.Headers)
		headersJSON = string(b)
	}

	// Store the request.
	req := &store.Request{
		Method:  incoming.Method,
		Path:    incoming.Path,
		Headers: headersJSON,
		Query:   incoming.Query,
		Body:    []byte(incoming.Body),
	}
	if err := h.Repo.AppendRequest(ctx, endpointID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store request")
		return
	}

	// Broadcast via SSE.
	broadcastRequest(endpointID, req)

	// Return the created request (with ID and timestamp).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// v1ListRequests handles GET /v1/endpoints/{endpointID}/requests
func (h *Handler) v1ListRequests(w http.ResponseWriter, r *http.Request) {
	endpointID := chi.URLParam(r, "endpointID")
	if endpointID == "" {
		writeError(w, http.StatusBadRequest, "missing endpoint ID")
		return
	}
	ctx := r.Context()

	// Ensure endpoint exists (or create).
	if _, err := h.Repo.GetEndpoint(ctx, endpointID); err != nil {
		h.Repo.CreateEndpoint(ctx, endpointID)
	}

	// Parse query parameters.
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	query := r.URL.Query().Get("q")

	var requests []*store.Request
	var err error
	if query != "" {
		requests, err = h.Repo.SearchRequests(ctx, endpointID, query, limit)
	} else {
		requests, err = h.Repo.ListRequests(ctx, endpointID, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list requests")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// v1ReplayRequest handles POST /v1/endpoints/{endpointID}/requests/{requestID}/replay
func (h *Handler) v1ReplayRequest(w http.ResponseWriter, r *http.Request) {
	// Reuse the existing replay logic but with JSON response.
	// We'll just call the same handler as the HTML endpoint.
	HandleReplay(h.Repo)(w, r)
}

// v1SetMock handles PUT /v1/endpoints/{endpointID}/mock
func (h *Handler) v1SetMock(w http.ResponseWriter, r *http.Request) {
	// Reuse the existing mock handler.
	HandleMock(h.Repo)(w, r)
}

// v1OpenAPI returns the OpenAPI 3.1 specification for the v1 API.
func (h *Handler) v1OpenAPI(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "Hatch API",
			"version":     "v1",
			"description": "JSON API for Hatch webhook capture, inspection, replay, and mocking.",
		},
		"servers": []map[string]string{
			{"url": "/"},
		},
		"paths": map[string]interface{}{
			"/v1/endpoints/{endpointID}/requests": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Capture a request",
					"operationId": "captureRequest",
					"parameters": []map[string]interface{}{
						{"name": "endpointID", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/IncomingRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Request captured",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/StoredRequest",
									},
								},
							},
						},
					},
				},
				"get": map[string]interface{}{
					"summary":     "List or search requests",
					"operationId": "listRequests",
					"parameters": []map[string]interface{}{
						{"name": "endpointID", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
						{"name": "limit", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 100}},
						{"name": "q", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Full-text search query"},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of requests",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/StoredRequest",
										},
									},
								},
							},
						},
					},
				},
			},
			"/v1/endpoints/{endpointID}/requests/{requestID}/replay": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Replay a captured request",
					"operationId": "replayRequest",
					"parameters": []map[string]interface{}{
						{"name": "endpointID", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
						{"name": "requestID", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ReplayRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Replay result",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ReplayResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/endpoints/{endpointID}/mock": map[string]interface{}{
				"put": map[string]interface{}{
					"summary":     "Set mock configuration",
					"operationId": "setMock",
					"parameters": []map[string]interface{}{
						{"name": "endpointID", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/MockConfig",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Mock set",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"IncomingRequest": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"method":  map[string]interface{}{"type": "string", "default": "POST"},
						"path":    map[string]interface{}{"type": "string", "default": "/"},
						"headers": map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{"type": "string"}},
						"query":   map[string]interface{}{"type": "string"},
						"body":    map[string]interface{}{"type": "string"},
					},
				},
				"StoredRequest": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          map[string]interface{}{"type": "string"},
						"endpoint_id": map[string]interface{}{"type": "string"},
						"method":      map[string]interface{}{"type": "string"},
						"path":        map[string]interface{}{"type": "string"},
						"headers":     map[string]interface{}{"type": "string"},
						"query":       map[string]interface{}{"type": "string"},
						"body":        map[string]interface{}{"type": "string", "format": "byte"},
						"created_at":  map[string]interface{}{"type": "string", "format": "date-time"},
					},
				},
				"ReplayRequest": map[string]interface{}{
					"type":     "object",
					"required": []string{"target_url"},
					"properties": map[string]interface{}{
						"target_url": map[string]interface{}{"type": "string", "format": "uri"},
					},
				},
				"ReplayResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status":  map[string]interface{}{"type": "integer"},
						"headers": map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{"type": "string"}},
						"body":    map[string]interface{}{"type": "string"},
						"error":   map[string]interface{}{"type": "string"},
					},
				},
				"MockConfig": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status":  map[string]interface{}{"type": "integer"},
						"headers": map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{"type": "string"}},
						"body":    map[string]interface{}{"type": "string", "format": "byte"},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}
