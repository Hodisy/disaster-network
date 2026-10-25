package httpapi

import (
	"net/http"
	"yafa/backend/core"
)

// Owner management remains accessible for stored jobs when the model capability
// is switched off. There is deliberately no generic SSH/run HTTP endpoint.
func (h Handler) codexRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agent/codex/jobs/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil || h.Agent.Codex == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		job, err := h.Agent.Codex.Read(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, core.CodexJobSummary(job))
	}))
	mux.HandleFunc("POST /v1/agent/codex/jobs/{id}/cancel", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil || h.Agent.Codex == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		job, err := h.Agent.Codex.Cancel(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 202, core.CodexJobSummary(job))
	}))
}
