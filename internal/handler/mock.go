package handler

import (
	"encoding/json"
	"net/http"

	"github.com/elfoundation/hatch/internal/store"
)

// HandleMock handles PUT /e/{endpointID}/mock.
func HandleMock(repo store.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpointID := r.PathValue("endpointID")
		if endpointID == "" {
			writeError(w, http.StatusBadRequest, "missing endpoint ID")
			return
		}
		if _, err := repo.GetEndpoint(r.Context(), endpointID); err != nil {
			repo.CreateEndpoint(r.Context(), endpointID)
		}
		var mock store.MockConfig
		if err := json.NewDecoder(r.Body).Decode(&mock); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
			return
		}
		mock.EndpointID = endpointID
		if err := repo.SetMock(r.Context(), &mock); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to set mock: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
