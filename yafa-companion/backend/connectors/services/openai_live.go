package services

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
	"yafa/backend/entities"
	"yafa/backend/services"
)

type OpenAILive struct {
	APIKey string
	Model  string
}

func (c OpenAILive) Configured() bool { return c.APIKey != "" && c.Model != "" }
func (c OpenAILive) Open(ctx context.Context) (services.LiveConnection, error) {
	if !c.Configured() {
		return nil, &entities.ModelFailure{Code: "live_not_configured"}
	}
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, EnableCompression: false}
	conn, response, err := dialer.DialContext(ctx, "wss://api.openai.com/v1/live/sessions", http.Header{"Authorization": []string{"Bearer " + c.APIKey}})
	if err != nil {
		code := "live_connection_failed"
		if response != nil {
			switch response.StatusCode {
			case 401:
				code = "openai_authentication_failed"
			case 403:
				code = "live_access_denied"
			case 404:
				code = "live_unavailable"
			case 429:
				code = "live_rate_limited"
			}
			if response.Body != nil {
				response.Body.Close()
			}
		}
		return nil, &entities.ModelFailure{Code: code}
	}
	conn.SetReadLimit(262144)
	return &liveSocket{conn: conn}, nil
}

type liveSocket struct {
	conn  *websocket.Conn
	write sync.Mutex
}

func (s *liveSocket) Send(ctx context.Context, data json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.write.Lock()
	defer s.write.Unlock()
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = s.conn.SetWriteDeadline(deadline)
	return s.conn.WriteMessage(websocket.TextMessage, data)
}
func (s *liveSocket) Receive(ctx context.Context) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = s.conn.SetReadDeadline(deadline)
	}
	typeID, data, err := s.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if typeID != websocket.TextMessage || !json.Valid(data) {
		return nil, &entities.ModelFailure{Code: "invalid_live_event", UsageUnknown: true}
	}
	return data, nil
}
func (s *liveSocket) Close() error { return s.conn.Close() }
