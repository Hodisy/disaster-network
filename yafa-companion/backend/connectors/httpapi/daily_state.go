package httpapi

import (
	"net/http"
	"time"
	"yafa/backend/core"
)

func (h Handler) dailyStateRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/daily-state", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Agent == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		var date *string
		if r.URL.Query().Has("date") {
			value := r.URL.Query().Get("date")
			date = &value
		}
		value, err := h.Agent.ReadDailyState(r.Context(), date, time.Now())
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, value)
	}))
}
