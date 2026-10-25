package core

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sync"
	"time"
	"yafa/backend/entities"
	"yafa/backend/repositories"
	"yafa/backend/services"
)

type CodexJobs struct {
	Repository repositories.CodexRepository
	Runner     services.CodexRunner
	Safety     *SafetyCore
	Root       context.Context
	RuntimeID  string
	MaxSeconds int
	mu         sync.Mutex
	active     map[string]activeRun
}

func NewCodexJobs(root context.Context, repo repositories.CodexRepository, runner services.CodexRunner, safety *SafetyCore, runtime string, seconds int) *CodexJobs {
	return &CodexJobs{Repository: repo, Runner: runner, Safety: safety, Root: root, RuntimeID: runtime, MaxSeconds: seconds, active: map[string]activeRun{}}
}
func (c *CodexJobs) Configured() bool {
	return c != nil && c.Repository != nil && c.Runner != nil && c.Runner.Configured() && c.Safety != nil && c.Root != nil && c.MaxSeconds >= 30 && c.MaxSeconds <= 600
}

var codexIDPattern = regexp.MustCompile(`^codex-[a-f0-9]{32}$`)

func ValidCodexID(id string) bool { return codexIDPattern.MatchString(id) }
func codexCommand(name string) bool {
	return name == "run_codex" || name == "read_codex" || name == "cancel_codex"
}
func CodexJobSummary(job entities.CodexJob) map[string]any {
	return map[string]any{"job_id": job.ID, "run_id": job.RunID, "thread_id": job.ThreadID, "status": job.Status, "created_at": job.CreatedAt, "deadline": job.Deadline, "finished_at": job.FinishedAt, "cancel_requested": job.CancelRequested, "result": job.Result, "error_code": job.ErrorCode, "billing": "Codex account usage; separate from OpenAI API costs. No USD estimate is available."}
}
func (c *AgentHarness) executeCodex(ctx context.Context, runID, threadID string, call entities.ToolCall) (json.RawMessage, error) {
	if !c.commandAllowed(call.Name) {
		return nil, cliError("command_not_allowed", "Choose an available command from commands.md.")
	}
	args, fingerprint, err := toolArguments(call)
	if err != nil {
		return nil, err
	}
	var job entities.CodexJob
	if call.Name == "run_codex" {
		var in struct {
			Task string `json:"task"`
		}
		if strictArguments(args, &in) != nil || !validText(in.Task, 3000) {
			return nil, cliError("invalid_codex_task", "Provide a nonempty task of at most 3000 UTF-8 bytes.")
		}
		workspace, _, e := c.Workspace.Read(ctx)
		if e != nil {
			return nil, e
		}
		if workspace.Lease == nil || workspace.Lease.RunID != runID || !workspace.Lease.ExpiresAt.After(time.Now()) {
			return nil, ErrAgentBusy
		}
		job, err = c.Codex.Start(ctx, runID, threadID, fingerprint, in.Task)
	} else {
		var in struct {
			JobID string `json:"job_id"`
		}
		if strictArguments(args, &in) != nil {
			return nil, ErrInvalidRecord
		}
		if call.Name == "cancel_codex" {
			job, err = c.Codex.Cancel(ctx, in.JobID)
		} else {
			job, err = c.Codex.Read(ctx, in.JobID)
		}
	}
	return encoded(CodexJobSummary(job)), err
}
func (c *CodexJobs) Start(ctx context.Context, runID, threadID, fingerprint, task string) (entities.CodexJob, error) {
	var zero entities.CodexJob
	if !c.Configured() {
		return zero, ErrAgentUnavailable
	}
	if err := c.Root.Err(); err != nil {
		return zero, err
	}
	if !validAgentID(runID, "run") || !validAgentID(threadID, "thread") || !validText(task, 3000) {
		return zero, ErrInvalidRecord
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := stableID("codex", runID+":"+fingerprint)
	old, _, err := c.Repository.LoadCodexJob(ctx, id)
	if err == nil {
		return old, nil
	} // Never replay a persisted task, including failures.
	if !errors.Is(err, ErrRecordNotFound) {
		return zero, err
	}
	index, version, err := c.Repository.LoadCodexIndex(ctx)
	if err != nil {
		return zero, err
	}
	for _, existing := range index.IDs {
		if existing == id {
			return zero, cliError("codex_outcome_unknown", "A prior intent exists. Do not automatically relaunch this task.")
		}
	}
	if len(c.active) > 0 || (index.ActiveID != "" && index.ActiveUntil.After(time.Now())) {
		return zero, cliError("codex_busy", "One remote job is active or its stop is unconfirmed. Read its status or wait for its deadline.")
	}
	if len(index.IDs) >= 128 {
		return zero, ErrAgentCapacity
	}
	if err = c.Safety.Check(ctx); err != nil {
		return zero, err
	}
	now := time.Now().UTC()
	job := entities.CodexJob{ID: id, RunID: runID, ThreadID: threadID, RuntimeID: c.RuntimeID, Task: task, Status: "queued", CreatedAt: now, Deadline: now.Add(time.Duration(c.MaxSeconds+20) * time.Second)}
	index.IDs = append(index.IDs, id)
	index.ActiveID = id
	index.ActiveUntil = job.Deadline.Add(15 * time.Second)
	if _, err = c.Repository.SaveCodexIndex(ctx, version, index); err != nil {
		return zero, err
	}
	if _, err = c.Repository.SaveCodexJob(ctx, "0", job); err != nil {
		return zero, err
	}
	budget, err := c.Safety.Reserve(ctx, id, "codex", 0, 0)
	if err != nil {
		c.finishUnstarted(id, err)
		return zero, err
	}
	job.BudgetKey = budget
	job.Status = "running"
	_, v, err := c.Repository.LoadCodexJob(ctx, id)
	if err == nil {
		_, err = c.Repository.SaveCodexJob(ctx, v, job)
	}
	if err != nil {
		c.finishUnstarted(id, err)
		return zero, err
	}
	jobCtx, cancel := context.WithTimeout(c.Root, time.Duration(c.MaxSeconds+20)*time.Second)
	active := activeRun{cancel: cancel, done: make(chan struct{})}
	c.active[id] = active
	go c.run(jobCtx, job, active)
	return job, nil
}

// Called under mu, only before any remote request has started.
func (c *CodexJobs) finishUnstarted(id string, cause error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	job, v, err := c.Repository.LoadCodexJob(ctx, id)
	if err == nil {
		now := time.Now().UTC()
		job.FinishedAt = &now
		job.Status = "failed"
		job.ErrorCode = agentErrorCode(cause)
		_, _ = c.Repository.SaveCodexJob(ctx, v, job)
	}
	_ = c.Safety.Settle(ctx, budgetLedgerKey, id, "not_started", 0, 0, 0, true, true, "codex_not_started")
	c.release(ctx, id)
}
func (c *CodexJobs) release(ctx context.Context, id string) {
	index, v, err := c.Repository.LoadCodexIndex(ctx)
	if err == nil && index.ActiveID == id {
		index.ActiveID = ""
		_, _ = c.Repository.SaveCodexIndex(ctx, v, index)
	}
}
func (c *CodexJobs) run(ctx context.Context, job entities.CodexJob, active activeRun) {
	defer close(active.done)
	defer active.cancel()
	var result entities.CodexResult
	err := c.Safety.Check(ctx)
	if err == nil {
		result, err = c.Runner.Run(ctx, entities.CodexRequest{JobID: job.ID, Task: job.Task, MaxSeconds: c.MaxSeconds})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	defer delete(c.active, job.ID)
	savedCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	current, v, readErr := c.Repository.LoadCodexJob(savedCtx, job.ID)
	if readErr != nil {
		return
	}
	now := time.Now().UTC()
	current.FinishedAt = &now
	if err != nil {
		current.Status = "interrupted"
		current.ErrorCode = "codex_remote_outcome_unknown"
		// SSH disconnect/cancellation is not proof that the remote process stopped.
	} else {
		current.Status = result.Status
		current.Result = &result
		current.ErrorCode = result.ErrorCode
	}
	current.BudgetSettled = c.Safety.Settle(savedCtx, job.BudgetKey, job.ID, current.Status, 0, 0, 0, true, false, current.ErrorCode) == nil
	if _, err = c.Repository.SaveCodexJob(savedCtx, v, current); err != nil {
		return
	}
	if err == nil && current.Status != "interrupted" {
		c.release(savedCtx, job.ID)
	}
}
func (c *CodexJobs) Read(ctx context.Context, id string) (entities.CodexJob, error) {
	if !ValidCodexID(id) {
		return entities.CodexJob{}, ErrInvalidRecord
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	job, v, err := c.Repository.LoadCodexJob(ctx, id)
	if err != nil {
		return job, err
	}
	if _, active := c.active[id]; !active && (job.Status == "running" || job.Status == "queued") {
		job.Status = "interrupted"
		job.ErrorCode = "codex_remote_outcome_unknown"
		now := time.Now().UTC()
		job.FinishedAt = &now
		if _, err = c.Repository.SaveCodexJob(ctx, v, job); err != nil {
			return job, err
		}
		// Keep the remote slot held until its original hard deadline. No replay.
	}
	return job, nil
}
func (c *CodexJobs) Cancel(ctx context.Context, id string) (entities.CodexJob, error) {
	if !ValidCodexID(id) {
		return entities.CodexJob{}, ErrInvalidRecord
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	job, v, err := c.Repository.LoadCodexJob(ctx, id)
	if err != nil {
		return job, err
	}
	if job.Status != "queued" && job.Status != "running" {
		return job, nil
	}
	job.CancelRequested = true
	if _, err = c.Repository.SaveCodexJob(ctx, v, job); err != nil {
		return job, err
	}
	if active, ok := c.active[id]; ok {
		active.cancel()
	}
	return job, nil
}
func (c *CodexJobs) CancelAll() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, active := range c.active {
		active.cancel()
	}
}
func (c *CodexJobs) Shutdown(ctx context.Context) {
	if c == nil {
		return
	}
	c.mu.Lock()
	wait := []chan struct{}{}
	for _, active := range c.active {
		active.cancel()
		wait = append(wait, active.done)
	}
	c.mu.Unlock()
	for _, done := range wait {
		select {
		case <-done:
		case <-ctx.Done():
			return
		}
	}
}
