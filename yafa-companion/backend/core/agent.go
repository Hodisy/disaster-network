package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"
	"yafa/backend/entities"
	"yafa/backend/repositories"
	"yafa/backend/services"
)

type activeRun struct {
	cancel context.CancelFunc
	done   chan struct{}
}
type AgentHarness struct {
	Repository repositories.AgentRepository
	Workspace  WorkspaceCore
	Model      services.AgentModel
	Hosted     services.HostedTools
	Features   ToolFeatures
	Codex      *CodexJobs
	Safety     *SafetyCore
	Files      repositories.FileRepository
	Limits     entities.AgentLimits
	Root       context.Context
	RuntimeID  string
	mu         sync.Mutex
	active     map[string]activeRun
}

func NewAgentHarness(root context.Context, repo repositories.AgentRepository, model services.AgentModel, limits entities.AgentLimits) *AgentHarness {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("runtime identity unavailable")
	}
	return &AgentHarness{Repository: repo, Workspace: WorkspaceCore{Repository: repo, Now: time.Now}, Model: model, Limits: limits, Root: root, RuntimeID: hex.EncodeToString(bytes), active: map[string]activeRun{}}
}
func (c *AgentHarness) Configured() bool { return c != nil && c.Model != nil && c.Model.Configured() }
func validOrigin(origin string) bool {
	return origin == "mobile" || origin == "desktop" || origin == "home" || origin == "cloud"
}
func (c *AgentHarness) CreateThread(ctx context.Context, eventID, origin string) (entities.Thread, error) {
	if !agentEventPattern.MatchString(eventID) || !validOrigin(origin) {
		return entities.Thread{}, ErrInvalidRecord
	}
	id := stableID("thread", eventID)
	old, _, err := c.Repository.LoadThread(ctx, id)
	if err == nil {
		if old.Origin != origin {
			return entities.Thread{}, ErrVersionConflict
		}
		return old, nil
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return entities.Thread{}, err
	}
	thread := entities.Thread{ID: id, Origin: origin, CreatedAt: time.Now().UTC(), Messages: []entities.Message{}}
	_, err = c.Repository.SaveThread(ctx, "0", thread)
	if errors.Is(err, ErrVersionConflict) {
		old, _, err = c.Repository.LoadThread(ctx, id)
		if err == nil && old.Origin != origin {
			err = ErrVersionConflict
		}
		return old, err
	}
	return thread, err
}
func (c *AgentHarness) Thread(ctx context.Context, id string) (entities.Thread, error) {
	if !validAgentID(id, "thread") {
		return entities.Thread{}, ErrInvalidRecord
	}
	value, _, err := c.Repository.LoadThread(ctx, id)
	return value, err
}

type SubmitRun struct {
	ThreadID string `json:"thread_id"`
	EventID  string `json:"event_id"`
	Message  string `json:"message"`
}

func (c *AgentHarness) Submit(ctx context.Context, in SubmitRun) (entities.AgentRun, error) {
	if c.Root.Err() != nil {
		return entities.AgentRun{}, c.Root.Err()
	}
	if !validAgentID(in.ThreadID, "thread") || !agentEventPattern.MatchString(in.EventID) || !validText(in.Message, 4096) {
		return entities.AgentRun{}, ErrInvalidRecord
	}
	id := stableID("run", in.ThreadID+":"+in.EventID)
	hash := stableID("input", in.Message)
	old, _, err := c.Repository.LoadRun(ctx, id)
	if err == nil {
		if old.InputHash != hash || old.ThreadID != in.ThreadID {
			return entities.AgentRun{}, ErrVersionConflict
		}
		return c.ReadRun(ctx, id)
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return entities.AgentRun{}, err
	}
	if !c.Configured() {
		return entities.AgentRun{}, ErrAgentUnavailable
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.active) != 0 {
		return entities.AgentRun{}, ErrAgentBusy
	}
	thread, threadVersion, err := c.Repository.LoadThread(ctx, in.ThreadID)
	if err != nil {
		return entities.AgentRun{}, err
	}
	if len(thread.Messages) > 38 {
		return entities.AgentRun{}, ErrAgentCapacity
	}
	now := time.Now().UTC()
	deadline := now.Add(time.Duration(c.Limits.MaxDurationSeconds) * time.Second)
	leaseUntil := deadline.Add(15 * time.Second)
	if err = c.Workspace.Lease(ctx, id, &leaseUntil); err != nil {
		return entities.AgentRun{}, err
	}
	run := entities.AgentRun{ID: id, ThreadID: thread.ID, EventID: in.EventID, InputHash: hash, RuntimeID: c.RuntimeID, Status: "queued", Model: c.Model.Name(), CreatedAt: now, Limits: c.Limits, Tools: []entities.ToolTrace{}, ContextMessageIDs: []string{}}
	if c.Safety != nil {
		run.BudgetDate, err = c.Safety.Reserve(ctx, run.ID, "run", run.Limits.MaxTokens, 0)
		if err != nil {
			_ = c.Workspace.Lease(ctx, id, nil)
			return entities.AgentRun{}, err
		}
	}
	version, err := c.Repository.SaveRun(ctx, "0", run)
	if err == nil {
		thread.Messages = append(thread.Messages, entities.Message{ID: stableID("message", id+":user"), Role: "user", Text: in.Message, RunID: id, CreatedAt: now})
		_, err = c.Repository.SaveThread(ctx, threadVersion, thread)
	}
	if err == nil {
		err = c.recordThreadActivity(ctx, thread, run)
	}
	if err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = c.Workspace.Lease(cleanup, id, nil)
		if c.Safety != nil && run.BudgetDate != "" {
			_ = c.Safety.Settle(cleanup, run.BudgetDate, run.ID, "not_started", 0, 0, 0, true, true)
		}
		if version != "" {
			run.Status = "failed"
			run.ErrorCode = agentErrorCode(err)
			_, _ = c.Repository.SaveRun(cleanup, version, run)
		}
		return entities.AgentRun{}, err
	}
	runCtx, cancel := context.WithDeadline(c.Root, deadline)
	done := make(chan struct{})
	c.active[id] = activeRun{cancel: cancel, done: done}
	go func() {
		defer func() { cancel(); c.mu.Lock(); delete(c.active, id); close(done); c.mu.Unlock() }()
		c.execute(runCtx, run, version, thread)
	}()
	return run, nil
}
func (c *AgentHarness) ReadRun(ctx context.Context, id string) (entities.AgentRun, error) {
	if !validAgentID(id, "run") {
		return entities.AgentRun{}, ErrInvalidRecord
	}
	run, _, err := c.Repository.LoadRun(ctx, id)
	if err != nil {
		return run, err
	}
	c.mu.Lock()
	_, active := c.active[id]
	c.mu.Unlock()
	if (run.Status == "queued" || run.Status == "running") && (!active || run.RuntimeID != c.RuntimeID) {
		// A final save may have raced this read; refresh before deriving interruption.
		latest, _, readErr := c.Repository.LoadRun(ctx, id)
		if readErr != nil {
			return run, readErr
		}
		run = latest
		if run.Status != "queued" && run.Status != "running" {
			return run, nil
		}
		// Derived status; never automatically replay an interrupted external action.
		run.Status = "interrupted"
		run.ErrorCode = "execution_not_active"
		run.FinalizationConfirmed = false
		run.EffectsMayHaveCommitted = true
	}
	return run, nil
}
func (c *AgentHarness) Cancel(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if active, ok := c.active[id]; ok {
		active.cancel()
		return true
	}
	return false
}
func (c *AgentHarness) Wait(ctx context.Context, id string) (entities.AgentRun, error) {
	c.mu.Lock()
	active, ok := c.active[id]
	c.mu.Unlock()
	if ok {
		select {
		case <-active.done:
		case <-ctx.Done():
			return entities.AgentRun{}, ctx.Err()
		}
	}
	return c.ReadRun(ctx, id)
}
func (c *AgentHarness) Shutdown(ctx context.Context) {
	if c.Codex != nil {
		c.Codex.Shutdown(ctx)
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
func clipText(value string, n int) string {
	if len(value) <= n {
		return value
	}
	for n > 0 && !utf8.ValidString(value[:n]) {
		n--
	}
	return value[:n]
}
func encoded(value any) json.RawMessage { data, _ := json.Marshal(value); return data }
func (c *AgentHarness) context(thread entities.Thread) ([]json.RawMessage, []string) {
	items := []json.RawMessage{}
	ids := []string{}
	size := 0
	for i := len(thread.Messages) - 1; i >= 0 && len(items) < 12; i-- {
		message := thread.Messages[i]
		raw := encoded(map[string]any{"role": message.Role, "content": message.Text})
		if size+len(raw) > 12000 {
			break
		}
		size += len(raw)
		items = append([]json.RawMessage{raw}, items...)
		ids = append([]string{message.ID}, ids...)
	}
	return items, ids
}
func (c *AgentHarness) execute(ctx context.Context, run entities.AgentRun, version string, thread entities.Thread) {
	var executionErr error
	defer func() {
		if recover() != nil {
			executionErr = errors.New("execution_panicked")
			run.EffectsMayHaveCommitted = true
		}
		if ctx.Err() != nil {
			executionErr = ctx.Err()
		}
		if executionErr != nil {
			run.ErrorCode = agentErrorCode(executionErr)
			run.Status = "failed"
			if errors.Is(executionErr, context.Canceled) {
				run.Status = "cancelled"
			}
			if errors.Is(executionErr, context.DeadlineExceeded) {
				run.Status = "timed_out"
			}
			if errors.Is(executionErr, ErrAgentLimit) {
				run.Status = "limit_reached"
			}
		}
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if run.Status == "completed" {
			value, v, err := c.Repository.LoadThread(cleanup, run.ThreadID)
			if err == nil && len(value.Messages) < 40 {
				value.Messages = append(value.Messages, entities.Message{ID: stableID("message", run.ID+":assistant"), Role: "assistant", Text: run.Output, RunID: run.ID, CreatedAt: time.Now().UTC()})
				_, err = c.Repository.SaveThread(cleanup, v, value)
			} else if err == nil {
				err = ErrAgentCapacity
			}
			if err != nil {
				run.Status = "failed"
				run.ErrorCode = agentErrorCode(err)
			}
		}
		now := time.Now().UTC()
		run.FinishedAt = &now
		run.FinalizationConfirmed = true
		if c.Safety != nil && run.BudgetDate != "" {
			run.BudgetSettlementConfirmed = c.Safety.Settle(cleanup, run.BudgetDate, run.ID, run.Status, run.Usage.InputTokens+run.Usage.OutputTokens, 0, run.Usage.KnownCostUSD, !run.Usage.Incomplete && (run.Status == "completed" || run.Turns == 0), !run.Usage.Incomplete && !run.Usage.AdditionalCostsUnpriced && (run.Status == "completed" || run.Turns == 0), run.ErrorCode) == nil
		}
		if _, latest, err := c.Repository.LoadRun(cleanup, run.ID); err == nil {
			version = latest
		}
		_, _ = c.Repository.SaveRun(cleanup, version, run)
		_ = c.Workspace.Lease(cleanup, run.ID, nil)
	}()
	run.Status = "running"
	prompt, err := c.BuildPrompt(ctx, thread, "chat", time.Now(), "")
	if err != nil {
		executionErr = err
		return
	}
	run.PromptHash, run.PromptDocuments = prompt.Hash, prompt.Documents
	if executionErr = c.savePrompt(ctx, run.ID, prompt); executionErr != nil {
		return
	}
	input, ids := c.context(thread)
	run.ContextMessageIDs = ids
	executionErr = c.persistRun(ctx, &version, run)
	if executionErr != nil {
		return
	}
	seen := map[string]int{}
	failures := 0
	for turn := 0; turn < c.Limits.MaxTurns; turn++ {
		if executionErr = ctx.Err(); executionErr != nil {
			return
		}
		remaining := c.Limits.MaxTokens - run.Usage.InputTokens - run.Usage.OutputTokens
		if remaining < 16 {
			run.StopReason = "token_budget"
			executionErr = ErrAgentLimit
			return
		}
		request := entities.ModelRequest{Instructions: prompt.Instructions, Input: input, Tools: AgentTools(), MaxOutputTokens: min(c.Limits.MaxOutputTokens, int(remaining))}
		inputBytes := len(encoded(request.Input)) + len(encoded(request.Tools)) + len(request.Instructions)
		if inputBytes > c.Limits.MaxInputBytes {
			run.StopReason = "context_bytes_limit"
			executionErr = ErrAgentLimit
			return
		}
		// Reserve a conservative byte-based input allowance plus framing overhead.
		// Actual provider usage remains authoritative; this is not a billing quote.
		outputAllowance := remaining - int64(inputBytes) - 1024
		if outputAllowance < 16 {
			run.StopReason = "token_budget"
			executionErr = ErrAgentLimit
			return
		}
		request.MaxOutputTokens = min(request.MaxOutputTokens, int(outputAllowance))
		turnKey := fmt.Sprintf("%s-turn-%d", run.ID, turn)
		executionErr = c.Repository.SaveArtifact(ctx, turnKey+"-intent", encoded(map[string]any{"run_id": run.ID, "model": run.Model, "input_hash": stableID("context", string(encoded(input))), "status": "requested"}))
		if executionErr != nil {
			return
		}
		if c.Safety != nil {
			if err := c.Safety.Check(ctx); err != nil {
				executionErr = err
				return
			}
		}
		reply, err := c.Model.Respond(ctx, request)
		run.Turns++
		if err != nil {
			var failure *entities.ModelFailure
			if errors.As(err, &failure) && failure.UsageUnknown {
				run.Usage.Incomplete = true
				run.Usage.EstimatedCostUSD = nil
			}
			executionErr = err
			return
		}
		if reply.Usage.EstimatedCostUSD != nil {
			run.Usage.KnownCostUSD += *reply.Usage.EstimatedCostUSD
		} else {
			run.Usage.AdditionalCostsUnpriced = true
			run.Usage.EstimatedCostUSD = nil
		}
		run.Usage.InputTokens += reply.Usage.InputTokens
		run.Usage.OutputTokens += reply.Usage.OutputTokens
		run.Usage.CachedTokens += reply.Usage.CachedTokens
		run.Usage.ReasoningTokens += reply.Usage.ReasoningTokens
		if reply.Usage.Incomplete {
			run.Usage.Incomplete = true
		}
		refreshUsageCost(&run.Usage)
		executionErr = c.Repository.SaveArtifact(ctx, turnKey+"-result", encoded(map[string]any{"response": reply, "usage": reply.Usage}))
		if executionErr != nil {
			return
		}
		executionErr = c.persistRun(ctx, &version, run)
		if executionErr != nil {
			return
		}
		if run.Usage.Incomplete {
			executionErr = errors.New("usage_unavailable")
			return
		}
		if run.Usage.InputTokens+run.Usage.OutputTokens > c.Limits.MaxTokens || reply.Status == "incomplete" {
			run.StopReason = "token_or_model_output_limit"
			executionErr = ErrAgentLimit
			return
		}
		if reply.Status != "completed" {
			executionErr = errors.New("model_response_failed")
			return
		}
		if len(reply.Calls) == 0 {
			if !validText(reply.Text, 32768) {
				executionErr = errors.New("empty_model_response")
				return
			}
			run.Output = clipText(reply.Text, 4096)
			run.Status = "completed"
			return
		}
		input = append(input, reply.Output...)
		for _, call := range reply.Calls {
			output, err := c.executeCommandBatch(ctx, &run, &version, call, seen, &failures)
			if err != nil {
				executionErr = err
				return
			}
			input = append(input, encoded(map[string]any{"type": "function_call_output", "call_id": call.CallID, "output": string(output)}))
		}
	}
	run.StopReason = "turn_limit"
	executionErr = ErrAgentLimit
}

func (c *AgentHarness) persistRun(ctx context.Context, version *string, run entities.AgentRun) error {
	next, err := c.Repository.SaveRun(ctx, *version, run)
	if err == nil {
		*version = next
	}
	return err
}

func (c *AgentHarness) CompleteTask(ctx context.Context, id, eventID, expected string, completed bool) (json.RawMessage, error) {
	if !validAgentID(id, "task") || !agentEventPattern.MatchString(eventID) {
		return nil, ErrInvalidRecord
	}
	in := CompleteTaskInput{ID: id, Expected: expected, Completed: completed}
	fingerprint := stableID("manual", string(encoded(in)))
	return c.Workspace.CompleteTask(ctx, "", stableID("action", "manual:"+eventID), fingerprint, in)
}

func (c *AgentHarness) CancelAll() {
	if c.Codex != nil {
		c.Codex.CancelAll()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, run := range c.active {
		run.cancel()
	}
}
