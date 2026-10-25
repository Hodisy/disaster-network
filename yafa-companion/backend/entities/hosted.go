package entities

import (
	"encoding/json"
	"time"
)

// Binary data never enters a prompt, trace or run record.
type GeneratedFile struct {
	Name string
	MIME string
	Data []byte
}
type StoredFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MIME      string    `json:"mime_type"`
	Size      int       `json:"size_bytes"`
	SHA256    string    `json:"sha256"`
	Chunks    int       `json:"chunks"`
	RunID     string    `json:"run_id"`
	ThreadID  string    `json:"thread_id"`
	CreatedAt time.Time `json:"created_at"`
}
type HostedRequest struct {
	Kind            string
	Instruction     string
	Quality         string
	MaxOutputTokens int
}
type HostedResult struct {
	AdditionalCostUSD *float64
	ID                string
	Model             string
	Text              string
	Files             []GeneratedFile
	Usage             Usage
	ProviderUsage     json.RawMessage
	ToolCalls         int
	ContainerIDs      []string
}
type HostedCall struct {
	AdditionalCostUSD *float64        `json:"estimated_additional_cost_usd"`
	ID                string          `json:"id"`
	Kind              string          `json:"kind"`
	Status            string          `json:"status"`
	ProviderID        string          `json:"provider_id,omitempty"`
	Model             string          `json:"model,omitempty"`
	Usage             Usage           `json:"usage"`
	ProviderUsage     json.RawMessage `json:"provider_usage,omitempty"`
	ToolCalls         int             `json:"hosted_tool_calls"`
	ContainerIDs      []string        `json:"container_ids,omitempty"`
	Files             []StoredFile    `json:"files,omitempty"`
	ErrorCode         string          `json:"error_code,omitempty"`
}
