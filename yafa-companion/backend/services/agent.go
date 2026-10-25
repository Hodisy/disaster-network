package services

import (
	"context"
	"encoding/json"
	"yafa/backend/entities"
)

type AgentModel interface {
	Configured() bool
	Name() string
	Respond(context.Context, entities.ModelRequest) (entities.ModelReply, error)
}
type LiveConnection interface {
	Send(context.Context, json.RawMessage) error
	Receive(context.Context) (json.RawMessage, error)
	Close() error
}
type LiveProvider interface {
	Configured() bool
	Open(context.Context) (LiveConnection, error)
}

// VoicePeer carries our app protocol, not arbitrary provider commands.
type VoicePeer interface {
	Receive(context.Context) (json.RawMessage, error)
	Send(context.Context, any) error
	Close() error
}
