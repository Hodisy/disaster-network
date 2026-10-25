package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"yafa/backend/core"
)

func agentHTTPError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrSafetyBlocked):
		writeJSON(w, 429, map[string]string{"error": "safety_paused_or_limit", "message": "AI is paused, disabled or its rolling quota is reached. Review Settings; no automatic quota increase occurs."})
	case errors.Is(err, core.ErrAgentUnavailable):
		writeJSON(w, 503, map[string]string{"error": "agent_not_configured", "message": "Configure private OPENAI_API_KEY and OPENAI_MODEL on the server."})
	case errors.Is(err, core.ErrAgentBusy):
		writeJSON(w, 409, map[string]string{"error": "agent_busy"})
	case errors.Is(err, core.ErrAgentCapacity):
		writeJSON(w, 409, map[string]string{"error": "capacity_reached", "message": "This initial bounded workspace or thread has reached its storage limit."})
	default:
		writeError(w, err)
	}
}
func (h Handler) agentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agent/status", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		writeJSON(w, 200, map[string]any{"configured": h.Agent.Configured(), "limits": h.Agent.Limits, "live_configured": h.Voice != nil && h.Voice.Configured(), "ui_sync_available": false, "provider_function": "command", "hosted_tools": h.Agent.HostedCapabilities(), "capabilities": h.Agent.Capabilities(), "registered_commands": h.Agent.RegisteredCommands()})
	}))
	mux.HandleFunc("POST /v1/threads", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			EventID string `json:"event_id"`
			Origin  string `json:"origin"`
		}
		if !decode(w, r, &input) {
			return
		}
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		thread, err := h.Agent.CreateThread(r.Context(), input.EventID, input.Origin)
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, thread)
	}))
	mux.HandleFunc("GET /v1/threads/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		thread, err := h.Agent.Thread(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, thread)
	}))
	mux.HandleFunc("POST /v1/agent/runs", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input core.SubmitRun
		if !decode(w, r, &input) {
			return
		}
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		run, err := h.Agent.Submit(r.Context(), input)
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, run)
	}))
	mux.HandleFunc("GET /v1/agent/runs/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		run, err := h.Agent.ReadRun(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, run)
	}))
	mux.HandleFunc("POST /v1/agent/runs/{id}/cancel", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		if _, err := h.Agent.ReadRun(r.Context(), r.PathValue("id")); err != nil {
			agentHTTPError(w, err)
			return
		}
		requested := h.Agent.Cancel(r.PathValue("id"))
		writeJSON(w, 202, map[string]any{"cancellation_requested": requested, "message": "Read the run to confirm its final status; already committed actions are not undone."})
	}))
	for _, kind := range []string{"projects", "tasks"} {
		kind := kind
		mux.HandleFunc("GET /v1/"+kind, h.protect(func(w http.ResponseWriter, r *http.Request) {
			if h.Agent == nil {
				agentHTTPError(w, core.ErrAgentUnavailable)
				return
			}
			offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
			if r.URL.Query().Get("offset") == "" {
				offset = 0
				err = nil
			}
			if err != nil {
				writeError(w, core.ErrInvalidRecord)
				return
			}
			var value any
			if kind == "projects" {
				value, err = h.Agent.Workspace.Projects(r.Context(), r.URL.Query().Get("include_archived") == "true", offset, 20)
			} else {
				var project *string
				if p := r.URL.Query().Get("project_id"); p != "" {
					project = &p
				}
				value, err = h.Agent.Workspace.Tasks(r.Context(), project, r.URL.Query().Get("include_completed") == "true", offset, 20)
			}
			if err != nil {
				agentHTTPError(w, err)
				return
			}
			writeJSON(w, 200, value)
		}))
	}
	mux.HandleFunc("PATCH /v1/tasks/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			EventID   string `json:"event_id"`
			Expected  string `json:"expected_revision"`
			Completed *bool  `json:"completed"`
		}
		if !decode(w, r, &input) {
			return
		}
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		if input.Completed == nil {
			writeError(w, core.ErrInvalidRecord)
			return
		}
		value, err := h.Agent.CompleteTask(r.Context(), r.PathValue("id"), input.EventID, input.Expected, *input.Completed)
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, value)
	}))
}
