package httpapi

import (
	"net/http"
	"yafa/backend/core"
)

func (h Handler) safetyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/safety", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil || h.Agent.Safety == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		status, err := h.Agent.Safety.Status(r.Context())
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		if h.Voice != nil {
			status["voice_max_seconds"] = h.Voice.MaxDurationSeconds
			status["voice_idle_seconds"] = h.Voice.IdleSeconds
			status["voice_usd_per_minute"] = h.Voice.USDPerMinute
		}
		status["pricing_as_of"] = "2026-10-25"
		status["pricing_source"] = "https://developers.openai.com/api/docs/pricing"
		writeJSON(w, 200, status)
	}))
	mux.HandleFunc("PATCH /v1/safety", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Expected  string              `json:"expected_revision"`
			Confirmed bool                `json:"user_confirmed"`
			Settings  core.SafetySettings `json:"settings"`
		}
		if !decode(w, r, &input) {
			return
		}
		if !input.Confirmed {
			writeJSON(w, 400, map[string]string{"error": "Explicit user confirmation is required to change AI limits."})
			return
		}
		if h.Agent == nil || h.Agent.Safety == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		if err := h.Agent.Safety.Update(r.Context(), input.Expected, input.Settings); err != nil {
			agentHTTPError(w, err)
			return
		}
		if input.Settings.Paused {
			h.Agent.CancelAll()
			if h.Voice != nil {
				h.Voice.Stop()
			}
		}
		writeJSON(w, 200, map[string]any{"saved": true, "paused": input.Settings.Paused, "active_stop_requested": input.Settings.Paused})
	}))
	mux.HandleFunc("POST /v1/agent/commands/validate", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Command string `json:"command"`
		}
		if !decode(w, r, &input) {
			return
		}
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		names, err := h.Agent.ValidateCommands(input.Command)
		if err != nil {
			writeJSON(w, 400, map[string]any{"valid": false, "executed": false, "message": core.CommandValidationMessage(err)})
			return
		}
		writeJSON(w, 200, map[string]any{"valid": true, "executed": false, "commands": names})
	}))
}
