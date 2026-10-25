package core

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

var (
	ErrMemoryUnavailable = errors.New("memory unavailable")
	ErrRecordNotFound    = errors.New("record not found")
	ErrVersionConflict   = errors.New("record version changed")
	ErrInvalidRecord     = errors.New("invalid record request")
)
var recordKey = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,95}$`)

type Memory struct{ Repository repositories.MemoryRepository }

func (c Memory) Ready(ctx context.Context) error {
	if c.Repository == nil {
		return ErrMemoryUnavailable
	}
	return c.Repository.Ready(ctx)
}
func (c Memory) Read(ctx context.Context, key string) (entities.MemoryRecord, error) {
	if !recordKey.MatchString(key) {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	if c.Repository == nil {
		return entities.MemoryRecord{}, ErrMemoryUnavailable
	}
	return c.Repository.Read(ctx, key)
}
func validObject(value json.RawMessage) bool {
	var object map[string]json.RawMessage
	return len(value) <= 262144 && json.Unmarshal(value, &object) == nil && object != nil
}
func (c Memory) Create(ctx context.Context, key string, value json.RawMessage) (entities.MemoryRecord, error) {
	if !recordKey.MatchString(key) || !validObject(value) {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	if c.Repository == nil {
		return entities.MemoryRecord{}, ErrMemoryUnavailable
	}
	return c.Repository.Write(ctx, key, "0", value)
}
func (c Memory) Patch(ctx context.Context, key, version string, changes json.RawMessage) (entities.MemoryRecord, error) {
	if !recordKey.MatchString(key) || !validObject(changes) {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	n, err := strconv.ParseUint(version, 10, 64)
	if err != nil || n == 0 {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	current, err := c.Read(ctx, key)
	if err != nil {
		return entities.MemoryRecord{}, err
	}
	if current.Version != version {
		return entities.MemoryRecord{}, ErrVersionConflict
	}
	var original, patch map[string]json.RawMessage
	if json.Unmarshal(current.Value, &original) != nil || original == nil || json.Unmarshal(changes, &patch) != nil {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	// Explicit top-level field replacement. Null is a stored null, not erasure.
	for field, value := range patch {
		original[field] = value
	}
	value, err := json.Marshal(original)
	if err != nil || !validObject(value) {
		return entities.MemoryRecord{}, ErrInvalidRecord
	}
	return c.Repository.Write(ctx, key, version, value)
}
