package httpapi

import (
	"net/http"
	"strconv"
	"time"
	"yafa/backend/core"
)

func (h Handler) agendaRoutes(mux *http.ServeMux) {
	for _, period := range []string{"today", "day", "week", "month", "range"} {
		period := period
		mux.HandleFunc("GET /v1/agenda/"+period, h.protect(func(w http.ResponseWriter, r *http.Request) {
			if h.Agent == nil {
				agentHTTPError(w, core.ErrAgentUnavailable)
				return
			}
			query := r.URL.Query()
			owner := query.Get("owner")
			if owner == "" {
				owner = "both"
			}
			in := core.AgendaInput{Owner: owner, IncludeCompleted: true, Limit: 20}
			if query.Has("include_completed") {
				value, err := strconv.ParseBool(query.Get("include_completed"))
				if err != nil {
					writeError(w, core.ErrInvalidRecord)
					return
				}
				in.IncludeCompleted = value
			}
			for name, destination := range map[string]*int{"offset": &in.Offset, "limit": &in.Limit} {
				if query.Has(name) {
					value, err := strconv.Atoi(query.Get(name))
					if err != nil {
						writeError(w, core.ErrInvalidRecord)
						return
					}
					*destination = value
				}
			}
			key := "date"
			if period == "week" {
				key = "week_of"
			}
			if period != "today" && query.Has(key) {
				value := query.Get(key)
				in.Date = &value
			}
			var value any
			var err error
			switch period {
			case "month":
				var month *string
				if query.Has("month") {
					raw := query.Get("month")
					month = &raw
				}
				value, err = h.Agent.ReadAgendaMonth(r.Context(), in, month, time.Now())
			case "range":
				value, err = h.Agent.ReadAgendaRange(r.Context(), in, query.Get("start_date"), query.Get("end_date"))
			default:
				value, err = h.Agent.ReadAgenda(r.Context(), in, period == "week", time.Now())
			}
			if err != nil {
				agentHTTPError(w, err)
				return
			}
			writeJSON(w, 200, value)
		}))
	}
}
