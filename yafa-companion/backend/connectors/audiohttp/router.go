// Package audiohttp exposes the standalone, loopback-only USB audio bench.
package audiohttp

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"yafa/backend/core"
)

type Handler struct {
	Core                    core.AudioLab
	Live                    *core.AudioLabLive
	Token                   string
	TranscriptionConfigured bool
}

func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func failure(w http.ResponseWriter, err error) {
	status, message := http.StatusBadGateway, "Audio conversion or provider unavailable"
	if errors.Is(err, core.ErrInvalidLabAudio) {
		status, message = 400, "Use an MP3 or WAV up to 2 MiB or a 16 kHz mono PCM16 WAV recording"
	}
	if errors.Is(err, core.ErrConsentRequired) {
		status, message = 403, "Explicit OpenAI transcription consent required"
	}
	if errors.Is(err, core.ErrAudioLabUnavailable) {
		message = err.Error()
	}
	reply(w, status, map[string]string{"error": message})
}
func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	h.liveRoute(mux)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]any{"status": "ready", "transcription_configured": h.TranscriptionConfigured, "live_configured": h.Live != nil && h.Live.Configured(), "transport": "usb", "maximum_playback_seconds": 15})
	})
	mux.HandleFunc("POST /v1/audio-lab/prepare", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 2097152)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			failure(w, core.ErrInvalidLabAudio)
			return
		}
		result, err := h.Core.Prepare(r.Context(), data)
		if err != nil {
			failure(w, err)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("POST /v1/audio-lab/transcriptions", func(w http.ResponseWriter, r *http.Request) {
		if !h.TranscriptionConfigured {
			reply(w, 503, map[string]string{"error": "Set OPENAI_API_KEY in backend/.env.audio-lab, then restart the audio lab"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 320080)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			failure(w, core.ErrInvalidLabAudio)
			return
		}
		result, err := h.Core.Transcribe(r.Context(), data, r.Header.Get("X-Yafa-Audio-Consent") == "send-to-openai")
		if err != nil {
			failure(w, err)
			return
		}
		reply(w, 200, result)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Browser requests use the CSRF-protected USB helper, never this private service.
		if r.Host != "127.0.0.1:4181" && r.Host != "localhost:4181" {
			reply(w, 403, map[string]string{"error": "Local audio service only"})
			return
		}
		if len(h.Token) < 32 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.Token)) != 1 {
			reply(w, 401, map[string]string{"error": "Audio lab credential required"})
			return
		}
		mux.ServeHTTP(w, r)
	})
}
