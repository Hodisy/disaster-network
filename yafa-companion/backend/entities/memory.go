package entities

import "encoding/json"

// MemoryRecord is an infrastructure record, not yet a task or project contract.
// Version stays a string to avoid losing uint64 precision in JavaScript clients.
type MemoryRecord struct {
	Key          string          `json:"key"`
	Branch       string          `json:"branch"`
	Version      string          `json:"version"`
	Value        json.RawMessage `json:"value"`
	Notification string          `json:"notification,omitempty"`
}
