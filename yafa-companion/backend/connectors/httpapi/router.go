package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"yafa/backend/core"
)

type Handler struct {
	Memory      core.Memory
	Token       string
	Device      core.DeviceLink
	DeviceToken string
	Agent       *core.AgentHarness
	Voice       *core.VoiceSessions
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	status, message := http.StatusServiceUnavailable, "Memory service unavailable"
	switch {
	case errors.Is(err, core.ErrInvalidRecord):
		status, message = http.StatusBadRequest, "Invalid record request"
	case errors.Is(err, core.ErrRecordNotFound):
		status, message = http.StatusNotFound, "Record not found"
	case errors.Is(err, core.ErrVersionConflict):
		status, message = http.StatusConflict, "Record changed; read its latest version"
	}
	writeJSON(w, status, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, input any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 524288)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(input); err != nil {
		writeError(w, core.ErrInvalidRecord)
		return false
	}
	var trailing any
	if d.Decode(&trailing) != io.EOF {
		writeError(w, core.ErrInvalidRecord)
		return false
	}
	return true
}
func (h Handler) protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.Token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
			return
		}
		next(w, r)
	}
}
func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("GET /health/ready", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if err := h.Memory.Ready(r.Context()); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "scope": "storage API; agent configuration is reported separately", "agent_configured": h.Agent != nil && h.Agent.Configured(), "ui_sync_available": false})
	}))
	h.deviceAdminRoutes(mux)
	h.agentRoutes(mux)
	h.agendaRoutes(mux)
	h.dailyStateRoutes(mux)
	h.fileRoutes(mux)
	h.safetyRoutes(mux)
	h.codexRoutes(mux)
	h.liveRoutes(mux)
	mux.HandleFunc("GET /v1/records/{key}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		result, err := h.Memory.Read(r.Context(), r.PathValue("key"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}))
	mux.HandleFunc("POST /v1/records/{key}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Value json.RawMessage `json:"value"`
		}
		if !decode(w, r, &input) {
			return
		}
		result, err := h.Memory.Create(r.Context(), r.PathValue("key"), input.Value)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, result)
	}))
	mux.HandleFunc("PATCH /v1/records/{key}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ExpectedVersion string          `json:"expected_version"`
			Changes         json.RawMessage `json:"changes"`
		}
		if !decode(w, r, &input) {
			return
		}
		result, err := h.Memory.Patch(r.Context(), r.PathValue("key"), input.ExpectedVersion, input.Changes)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}))
	for _, route := range []string{"QUERY /v1/records", "DELETE /v1/records/{key}", "GET /v1/events"} {
		mux.HandleFunc(route, h.protect(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "This capability is not implemented in the local infrastructure slice"})
		}))
	}
	return mux
}
