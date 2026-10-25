package repositories

import (
	"context"
	"encoding/json"
	"yafa/backend/entities"
)

// The memory owner serializes compare-and-write; no invented fallback is allowed.
type MemoryRepository interface {
	Ready(context.Context) error
	Read(context.Context, string) (entities.MemoryRecord, error)
	Write(context.Context, string, string, json.RawMessage) (entities.MemoryRecord, error)
}
