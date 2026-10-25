package entities

import "time"

// Protocol shared by the SSH transport and the fixed remote runner. No shell
// command, filesystem path or callback URL comes from model arguments.
type CodexRequest struct {
	JobID      string `json:"job_id"`
	Task       string `json:"task"`
	MaxSeconds int    `json:"max_seconds"`
}
type CodexUsage struct {
	InputTokens       int64 `json:"input_tokens"`
	CachedInputTokens int64 `json:"cached_input_tokens"`
	OutputTokens      int64 `json:"output_tokens"`
	Confirmed         bool  `json:"confirmed"`
}
type CodexResult struct {
	Status       string     `json:"status"`
	Text         string     `json:"text"`
	Usage        CodexUsage `json:"usage"`
	ChangedFiles []string   `json:"remote_changed_files"`
	ErrorCode    string     `json:"error_code,omitempty"`
}
type CodexEvent struct {
	Type   string       `json:"type"`
	Phase  string       `json:"phase,omitempty"`
	Result *CodexResult `json:"result,omitempty"`
}
type CodexJob struct {
	ID              string       `json:"id"`
	RunID           string       `json:"run_id"`
	ThreadID        string       `json:"thread_id"`
	RuntimeID       string       `json:"runtime_id"`
	Task            string       `json:"task"`
	Status          string       `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	Deadline        time.Time    `json:"deadline"`
	FinishedAt      *time.Time   `json:"finished_at,omitempty"`
	CancelRequested bool         `json:"cancel_requested"`
	Result          *CodexResult `json:"result,omitempty"`
	ErrorCode       string       `json:"error_code,omitempty"`
	BudgetKey       string       `json:"budget_key,omitempty"`
	BudgetSettled   bool         `json:"budget_settled"`
}
type CodexIndex struct {
	IDs         []string  `json:"ids"`
	ActiveID    string    `json:"active_id,omitempty"`
	ActiveUntil time.Time `json:"active_until"`
}
