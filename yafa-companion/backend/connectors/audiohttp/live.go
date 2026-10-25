package audiohttp

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"yafa/backend/entities"
)

type livePeer struct{ conn *websocket.Conn }

func (p *livePeer) Receive(ctx context.Context) (json.RawMessage, error) {
	_, data, err := p.conn.ReadMessage()
	return data, err
}
func (p *livePeer) Send(ctx context.Context, value any) error {
	_ = p.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	return p.conn.WriteJSON(value)
}
func (p *livePeer) Close() error { return p.conn.Close() }
func (h Handler) liveRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/audio-lab/live", func(w http.ResponseWriter, r *http.Request) {
		if h.Live == nil || !h.Live.Configured() {
			reply(w, 503, map[string]string{"error": "OpenAI Live is not configured"})
			return
		}
		upgrade := websocket.Upgrader{ReadBufferSize: 8192, WriteBufferSize: 8192, CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == "" }}
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetReadLimit(16384)
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var start struct {
			Type          string `json:"type"`
			UserInitiated bool   `json:"user_initiated"`
		}
		if conn.ReadJSON(&start) != nil || start.Type != "session.start" || !start.UserInitiated {
			return
		}
		_ = conn.SetReadDeadline(time.Time{})
		peer := &livePeer{conn: conn}
		if err = h.Live.Run(r.Context(), peer, start.UserInitiated); err != nil {
			code := "live_session_failed"
			if failure, ok := err.(*entities.ModelFailure); ok {
				code = failure.Code
			}
			_ = peer.Send(context.Background(), map[string]any{"type": "voice.error", "error": code})
		}
	})
}
