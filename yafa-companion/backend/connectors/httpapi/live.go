package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
	"yafa/backend/core"
)

type voicePeer struct {
	conn  *websocket.Conn
	write sync.Mutex
}

func (p *voicePeer) Receive(ctx context.Context) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	kind, data, err := p.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if kind != websocket.TextMessage {
		return nil, core.ErrInvalidRecord
	}
	return data, nil
}
func strictClientVoiceStart(data []byte, value *core.VoiceStart) error {
	if len(data) > 2048 {
		return core.ErrInvalidRecord
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil || d.Decode(new(any)) != io.EOF {
		return core.ErrInvalidRecord
	}
	return nil
}
func (p *voicePeer) Send(ctx context.Context, value any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.write.Lock()
	defer p.write.Unlock()
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = p.conn.SetWriteDeadline(deadline)
	return p.conn.WriteJSON(value)
}
func (p *voicePeer) Close() error { return p.conn.Close() }
func (h Handler) liveRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/live/sessions/{id}", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Voice == nil {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		value, err := h.Voice.Read(r.Context(), r.PathValue("id"))
		if err != nil {
			agentHTTPError(w, err)
			return
		}
		writeJSON(w, 200, value)
	}))
	mux.HandleFunc("GET /v1/live/stream", h.protect(func(w http.ResponseWriter, r *http.Request) {
		if h.Voice == nil || !h.Voice.Configured() {
			agentHTTPError(w, core.ErrAgentUnavailable)
			return
		}
		upgrader := websocket.Upgrader{ReadBufferSize: 4096, WriteBufferSize: 4096, HandshakeTimeout: 5 * time.Second, CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			u, err := url.Parse(origin)
			return err == nil && u.Host == r.Host && (u.Scheme == "http" || u.Scheme == "https")
		}}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetReadLimit(32768)
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var start core.VoiceStart
		kind, data, err := conn.ReadMessage()
		if err != nil || kind != websocket.TextMessage {
			return
		}
		if strictClientVoiceStart(data, &start) != nil {
			_ = conn.WriteJSON(map[string]string{"type": "voice.error", "error": "invalid_session_start"})
			return
		}
		_ = conn.SetReadDeadline(time.Time{})
		peer := &voicePeer{conn: conn}
		if err = h.Voice.Run(r.Context(), peer, start); err != nil {
			_ = peer.Send(context.Background(), map[string]string{"type": "voice.error", "error": "voice_session_ended", "message": "Read the saved session/run state before retrying any action."})
		}
	}))
}
