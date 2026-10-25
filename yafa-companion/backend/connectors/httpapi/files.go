package httpapi

import (
	"mime"
	"net/http"
	"strconv"
	"yafa/backend/core"
)

// Owner-authenticated downloads only. No public URLs, raw host paths or device
// token access. Native clients fetch these bytes with the existing API token.
func (h Handler) fileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/files/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		meta, err := h.Agent.FileMetadata(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, meta)
	}))
	mux.HandleFunc("GET /v1/files/{id}/content", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		meta, data, err := h.Agent.FileContent(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		w.Header().Set("Content-Type", meta.MIME)
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": meta.Name}))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
}
