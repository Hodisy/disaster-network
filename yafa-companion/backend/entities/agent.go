package entities

import (
	"encoding/json"
	"time"
)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Revision    string `json:"revision"`
	Archived    bool   `json:"archived"`
}
type Task struct {
	ID             string     `json:"id"`
	ProjectID      *string    `json:"project_id"`
	Title          string     `json:"title"`
	Notes          string     `json:"notes"`
	DueAt          *time.Time `json:"due_at"`
	ScheduledDate  *string    `json:"scheduled_date"`
	ScheduledTime  *string    `json:"scheduled_time"`
	Completed      bool       `json:"completed"`
	Revision       string     `json:"revision"`
	SourceThreadID string     `json:"source_thread_id"`
}
type ActionReceipt struct {
	Fingerprint string          `json:"fingerprint"`
	Result      json.RawMessage `json:"result"`
}
type RunLease struct {
	RunID     string    `json:"run_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Initial bounded aggregate: entity changes and their action receipt commit together.
type Workspace struct {
	ExtraFields       map[string]json.RawMessage       `json:"-"`
	Projects          map[string]Project               `json:"projects"`
	Tasks             map[string]Task                  `json:"tasks"`
	Actions           map[string]ActionReceipt         `json:"actions"`
	ContextFiles      map[string]ContextFile           `json:"context_files,omitempty"`
	Lease             *RunLease                        `json:"lease,omitempty"`
	PersonalDocuments map[string]PersonalDocument      `json:"personal_documents,omitempty"`
	PersonalImports   map[string]PersonalImportReceipt `json:"personal_imports,omitempty"`
}
type ContextFile struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Revision string `json:"revision"`
	Source   string `json:"source"`
}
type PromptDocumentRef struct {
	Name     string `json:"name"`
	Revision string `json:"revision"`
	Source   string `json:"source"`
	Hash     string `json:"hash"`
}
type PromptSnapshot struct {
	Instructions string              `json:"instructions"`
	Documents    []PromptDocumentRef `json:"documents"`
	CreatedAt    time.Time           `json:"created_at"`
	Hash         string              `json:"hash"`
}

// One real thread activity per Paris date, regardless of the number of runs.
type ThreadActivity struct {
	ThreadID  string    `json:"thread_id"`
	Title     string    `json:"title"`
	Origin    string    `json:"origin"`
	LastRunID string    `json:"last_run_id"`
	UpdatedAt time.Time `json:"updated_at"`
}
type DayActivity struct {
	Date    string                    `json:"date"`
	Threads map[string]ThreadActivity `json:"threads"`
}
type StateProvenance struct {
	Quote      string    `json:"quote"`
	ThreadID   string    `json:"thread_id"`
	RunID      string    `json:"run_id"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsEstimate bool      `json:"is_estimate"`
}
type DailyGauge struct {
	Value      int             `json:"value"`
	Provenance StateProvenance `json:"provenance"`
}
type DailyWeather struct {
	Value      string          `json:"value"`
	Provenance StateProvenance `json:"provenance"`
}
type DailyState struct {
	Date     string                   `json:"date"`
	Revision string                   `json:"revision"`
	Weather  *DailyWeather            `json:"weather"`
	Gauges   map[string]DailyGauge    `json:"gauges"`
	Actions  map[string]ActionReceipt `json:"actions"`
}
type AgendaEntry struct {
	ID         string          `json:"id"`
	Owner      string          `json:"owner"`
	Kind       string          `json:"kind"`
	Date       string          `json:"date"`
	Time       *string         `json:"time"`
	DateSource string          `json:"date_source"`
	Task       *Task           `json:"task,omitempty"`
	Thread     *ThreadActivity `json:"thread,omitempty"`
}
type Agenda struct {
	StartDate    string        `json:"start_date"`
	EndDate      string        `json:"end_date"`
	Timezone     string        `json:"timezone"`
	Owner        string        `json:"owner"`
	Entries      []AgendaEntry `json:"entries"`
	Total        int           `json:"total"`
	NextOffset   *int          `json:"next_offset"`
	UndatedTasks int           `json:"undated_task_count"`
	Coverage     string        `json:"coverage"`
}
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	RunID     string    `json:"run_id"`
	CreatedAt time.Time `json:"created_at"`
}
type Thread struct {
	ID        string    `json:"id"`
	Origin    string    `json:"origin"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
}
type AgentLimits struct {
	MaxTurns           int   `json:"max_turns"`
	MaxToolCalls       int   `json:"max_tool_calls"`
	MaxTokens          int64 `json:"max_tokens"`
	MaxOutputTokens    int   `json:"max_output_tokens"`
	MaxInputBytes      int   `json:"max_input_bytes"`
	MaxDurationSeconds int   `json:"max_duration_seconds"`
}
type Usage struct {
	KnownCostUSD            float64  `json:"known_estimated_cost_usd"`
	AdditionalCostsUnpriced bool     `json:"additional_costs_unpriced,omitempty"`
	InputTokens             int64    `json:"input_tokens"`
	OutputTokens            int64    `json:"output_tokens"`
	CachedTokens            int64    `json:"cached_tokens"`
	ReasoningTokens         int64    `json:"reasoning_tokens"`
	Incomplete              bool     `json:"incomplete"`
	EstimatedCostUSD        *float64 `json:"estimated_cost_usd"`
}
type ToolTrace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	ErrorCode string `json:"error_code,omitempty"`
}
type AgentRun struct {
	BudgetDate                string              `json:"budget_date,omitempty"`
	BudgetSettlementConfirmed bool                `json:"budget_settlement_confirmed"`
	HostedCalls               []HostedCall        `json:"hosted_calls,omitempty"`
	ID                        string              `json:"id"`
	ThreadID                  string              `json:"thread_id"`
	EventID                   string              `json:"event_id"`
	InputHash                 string              `json:"input_hash"`
	RuntimeID                 string              `json:"runtime_id"`
	Status                    string              `json:"status"`
	Model                     string              `json:"model"`
	CreatedAt                 time.Time           `json:"created_at"`
	FinishedAt                *time.Time          `json:"finished_at,omitempty"`
	Limits                    AgentLimits         `json:"limits"`
	Turns                     int                 `json:"turns"`
	Tools                     []ToolTrace         `json:"tools"`
	Usage                     Usage               `json:"usage"`
	Output                    string              `json:"output"`
	ErrorCode                 string              `json:"error_code,omitempty"`
	StopReason                string              `json:"stop_reason,omitempty"`
	EffectsMayHaveCommitted   bool                `json:"effects_may_have_committed"`
	FinalizationConfirmed     bool                `json:"finalization_confirmed"`
	ContextMessageIDs         []string            `json:"context_message_ids"`
	PromptHash                string              `json:"prompt_hash,omitempty"`
	PromptDocuments           []PromptDocumentRef `json:"prompt_documents,omitempty"`
}
type ModelTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
	Strict      bool           `json:"strict"`
}
type ModelRequest struct {
	Instructions    string
	Input           []json.RawMessage
	Tools           []ModelTool
	MaxOutputTokens int
}
type ToolCall struct {
	CallID    string
	Name      string
	Arguments string
}
type ModelReply struct {
	ID     string            `json:"id"`
	Status string            `json:"status"`
	Output []json.RawMessage `json:"output"`
	Calls  []ToolCall        `json:"-"`
	Text   string            `json:"-"`
	Usage  Usage             `json:"-"`
}
type ModelFailure struct {
	Code         string
	UsageUnknown bool
}

func (e *ModelFailure) Error() string { return e.Code }

type VoiceConfig struct {
	Model        string           `json:"model"`
	Instructions string           `json:"instructions"`
	Audio        map[string]any   `json:"audio"`
	Delegation   map[string]any   `json:"delegation"`
	Input        []map[string]any `json:"input,omitempty"`
}
type VoiceTranscript struct {
	Role    string  `json:"role"`
	Text    string  `json:"text"`
	StartMS float64 `json:"start_ms"`
	EndMS   float64 `json:"end_ms"`
}
type VoiceRecord struct {
	BudgetDate                string              `json:"budget_date,omitempty"`
	BudgetSettlementConfirmed bool                `json:"budget_settlement_confirmed"`
	WallSeconds               float64             `json:"wall_seconds"`
	EstimatedCostUSD          *float64            `json:"estimated_voice_cost_usd"`
	CloseReason               string              `json:"close_reason,omitempty"`
	ID                        string              `json:"id"`
	ThreadID                  string              `json:"thread_id"`
	Surface                   string              `json:"surface"`
	Model                     string              `json:"model"`
	Status                    string              `json:"status"`
	StartedAt                 time.Time           `json:"started_at"`
	FinishedAt                *time.Time          `json:"finished_at,omitempty"`
	DurationSeconds           float64             `json:"duration_seconds"`
	Finalized                 bool                `json:"finalized"`
	UsageFinalConfirmed       bool                `json:"usage_final_confirmed"`
	ErrorCode                 string              `json:"error_code,omitempty"`
	Transcripts               []VoiceTranscript   `json:"transcripts"`
	RunIDs                    []string            `json:"run_ids"`
	PromptHash                string              `json:"prompt_hash,omitempty"`
	PromptDocuments           []PromptDocumentRef `json:"prompt_documents,omitempty"`
}
