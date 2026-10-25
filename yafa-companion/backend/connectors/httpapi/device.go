package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"yafa/backend/core"
	"yafa/backend/entities"
)

func decodeDevice(w http.ResponseWriter, r *http.Request, input any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(input) != nil || d.Decode(new(any)) != io.EOF {
		writeError(w, core.ErrInvalidRecord)
		return false
	}
	return true
}
func deviceError(w http.ResponseWriter, err error) {
	if errors.Is(err, core.ErrDeviceAccessDenied) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Device access denied"})
		return
	}
	if errors.Is(err, core.ErrNotConfigured) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Device service unavailable"})
		return
	}
	writeError(w, err)
}
func (h Handler) deviceAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/devices/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		link, online, err := h.Device.Read(r.Context(), r.PathValue("id"))
		if err != nil {
			deviceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"device": link, "online": online, "audio_stream_available": false})
	}))
	mux.HandleFunc("PUT /v1/devices/{id}/presentation", h.protect(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Expected   string `json:"expected_version"`
			Expression string `json:"expression"`
			Awake      *bool  `json:"awake"`
		}
		if !decodeDevice(w, r, &input) {
			return
		}
		if input.Awake == nil {
			writeError(w, core.ErrInvalidRecord)
			return
		}
		command, err := h.Device.SetPresentation(r.Context(), r.PathValue("id"), input.Expected, input.Expression, *input.Awake)
		if err != nil {
			deviceError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"desired": command, "status": "requested"})
	}))
}

// This listener deliberately exposes no infrastructure/admin record routes.
func (h Handler) DeviceRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/device/sync", func(w http.ResponseWriter, r *http.Request) {
		actual := sha256.Sum256([]byte(r.Header.Get("Authorization")))
		expected := sha256.Sum256([]byte("Bearer " + h.DeviceToken))
		if h.DeviceToken == "" || subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Device authentication required"})
			return
		}
		var report entities.DeviceReport
		if !decodeDevice(w, r, &report) {
			return
		}
		result, err := h.Device.Exchange(r.Context(), report)
		if err != nil {
			deviceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	return mux
}
