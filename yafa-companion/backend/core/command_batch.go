package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"yafa/backend/entities"
)

type commandFailure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type commandResult struct {
	Index   int             `json:"index"`
	Command string          `json:"command,omitempty"`
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *commandFailure `json:"error,omitempty"`
}

// Never expose raw storage/provider errors or user-supplied argument values.
func commandFailureFor(err error, name string) *commandFailure {
	code := agentErrorCode(err)
	var cli *commandError
	if errors.As(err, &cli) {
		return &commandFailure{Code: code, Message: cli.Hint}
	}
	message := "Execution unavailable. Stop; an action may have committed. Read actual state before any retry."
	switch code {
	case "not_found":
		message = "Record not found. Read the relevant list and use a returned ID."
	case "version_conflict":
		message = "Revision changed. Read the record again, then use its current revision."
	case "invalid_arguments":
		message = "Arguments violate domain rules. Check commands.md and read current state."
		switch name {
		case "get_day", "get_week", "get_month", "get_range", "read_daily_state":
			message = "Invalid date or page. Check date selectors, range order (at most 31 days), offset and limit."
		case "create_task":
			message = "Invalid task. Check the title, project ID, YYYY-MM-DD day, HH:mm time and RFC3339 deadline."
		case "complete_task", "update_project":
			message = "Invalid ID or revision. Read the task/project first; use the returned ID and revision."
		case "read_context_file", "update_context_file":
			message = "Use an allowed context name and valid content. For edits, read its revision first; preserve known {{variables}}."
		case "record_daily_state":
			message = "Check date, revision and 1–5 scores. Copy a direct quote from the current user message (at most 300 bytes)."
		case "create_project":
			message = "Check the project name and kind (personal or professional)."
		case "list_projects", "list_tasks":
			message = "Check IDs and pagination: offset starts at 0; limit is 1–20."
		}
	case "hosted_not_configured":
		message = "Hosted tool unavailable. Configure the server API key and model; no generation was confirmed."
	case "hosted_request_rejected", "hosted_resource_unavailable":
		message = "OpenAI rejected this hosted request or resource. Check model/account access; do not repeat it blindly."
	case "image_result_too_large", "analysis_files_too_large":
		message = "Generated files exceed the configured size limit. No complete downloadable result is confirmed."
	case "unsupported_analysis_file", "invalid_image_result":
		message = "Generated output failed file validation. No complete downloadable result is confirmed."
	case "analysis_incomplete", "analysis_execution_failed", "analysis_empty_result":
		message = "Hosted analysis did not complete successfully. No completed analysis is confirmed."
	case "hosted_outcome_unconfirmed":
		message = "An earlier hosted request has an unconfirmed outcome. Stop; do not bill it again."
	case "tool_not_allowed":
		message = "Call command with a documented CLI name from commands.md."
	case "capacity_reached":
		message = "Storage capacity reached. Stop and ask the user to review stored data."
	case "cancelled":
		message = "Execution cancelled. Do not assume the action was rolled back."
	case "deadline_exceeded":
		message = "Execution timed out. Read actual state before retrying an action."
	case "limit_reached":
		message = "Run limit reached. Stop issuing commands."
	}
	return &commandFailure{Code: code, Message: message}
}

func recoverableCommandError(err error) bool {
	return errors.Is(err, ErrInvalidRecord) || errors.Is(err, ErrRecordNotFound) || errors.Is(err, ErrVersionConflict) || (err != nil && err.Error() == "tool_not_allowed")
}

// Persist each command intent before execution and its actual outcome afterward.
// A validation rejection also consumes one tool slot, preventing free retries.
func (c *AgentHarness) commandStep(ctx context.Context, run *entities.AgentRun, version *string, provider entities.ToolCall, call entities.ToolCall, index int, validationErr error) (commandResult, error, error) {
	name := commandName(call.Name)
	if name == "" {
		name = "command"
	}
	stepID := stableID("step", fmt.Sprintf("%s:%s:%d", run.ID, provider.CallID, index))
	result := commandResult{Index: index, Command: name, Status: "requested"}
	err := c.Repository.SaveArtifact(ctx, stepID+"-intent", encoded(map[string]any{"run_id": run.ID, "call_id": provider.CallID, "command_index": index, "provider_function": provider.Name, "command": name, "arguments": clipText(call.Arguments, 8192)}))
	if err != nil {
		return result, nil, err
	}
	trace := entities.ToolTrace{ID: stepID, Name: name, Status: "requested"}
	run.Tools = append(run.Tools, trace)
	if validationErr == nil && !readOnlyAgentTool(call.Name) {
		run.EffectsMayHaveCommitted = true
	}
	if err = c.persistRun(ctx, version, *run); err != nil {
		return result, nil, err
	}
	var data json.RawMessage
	toolErr := validationErr
	if toolErr == nil {
		duration := 10 * time.Second
		if hostedCommand(call.Name) {
			duration = 75 * time.Second
		}
		toolCtx, cancel := context.WithTimeout(ctx, duration)
		if hostedCommand(call.Name) {
			data, toolErr = c.ExecuteHostedTool(toolCtx, run, version, call)
		} else {
			data, _, toolErr = c.ExecuteTool(toolCtx, run.ID, run.ThreadID, call)
		}
		cancel()
	}
	if toolErr != nil {
		trace.Status = "failed"
		trace.ErrorCode = agentErrorCode(toolErr)
		result.Status = "failed"
		if validationErr != nil {
			result.Status = "rejected"
		}
		result.Error = commandFailureFor(toolErr, call.Name)
	} else {
		trace.Status = "succeeded"
		result.Status = "succeeded"
		result.Data = data
	}
	if err = c.Repository.SaveArtifact(ctx, stepID+"-result", encoded(map[string]any{"run_id": run.ID, "result": result})); err != nil {
		return result, nil, err
	}
	run.Tools[len(run.Tools)-1] = trace
	if err = c.persistRun(ctx, version, *run); err != nil {
		return result, nil, err
	}
	if ctx.Err() != nil {
		return result, nil, ctx.Err()
	}
	return result, toolErr, nil
}

func (c *AgentHarness) executeCommandBatch(ctx context.Context, run *entities.AgentRun, version *string, provider entities.ToolCall, seen map[string]int, failures *int) (json.RawMessage, error) {
	prepared, failedAt, parseErr := parseCommandBatch(provider, c.CommandSpecs())
	required := len(prepared)
	for _, item := range prepared {
		if item.Call.Name == "generate_image" {
			required++
		}
		if item.Call.Name == "run_analysis" {
			required += 2
		}
	}
	if parseErr != nil {
		required = 1
	}
	if usedToolSlots(*run)+required > c.Limits.MaxToolCalls {
		run.StopReason = "tool_call_limit"
		return nil, ErrAgentLimit
	}
	if parseErr != nil {
		result, err, storageErr := c.commandStep(ctx, run, version, provider, provider, failedAt, parseErr)
		if storageErr != nil {
			return nil, storageErr
		}
		if err != nil && !recoverableCommandError(err) {
			return nil, err
		}
		(*failures)++
		if *failures >= 3 {
			run.StopReason = "tool_error_limit"
			return nil, ErrAgentLimit
		}
		return encoded(map[string]any{"ok": false, "rejected": true, "executed": 0, "succeeded": 0, "results": []commandResult{result}}), nil
	}
	// Preflight repetitions too: reject an over-budget chain before side effects.
	pending := map[string]int{}
	for _, item := range prepared {
		pending[item.Fingerprint]++
		if seen[item.Fingerprint]+pending[item.Fingerprint] > 2 {
			run.StopReason = "repeated_tool_call"
			return nil, ErrAgentLimit
		}
	}
	results := []commandResult{}
	executed, succeeded := 0, 0
	var stopErr error
	for i, item := range prepared {
		if usedToolSlots(*run) >= c.Limits.MaxToolCalls {
			run.StopReason = "tool_call_limit"
			return nil, ErrAgentLimit
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		seen[item.Fingerprint]++
		result, err, storageErr := c.commandStep(ctx, run, version, provider, item.Call, i+1, nil)
		if storageErr != nil {
			result.Status = "unconfirmed"
			result.Error = &commandFailure{Code: "storage_or_execution_unavailable", Message: "Outcome could not be confirmed. Stop and read actual state before retrying."}
			err = storageErr
		}
		results = append(results, result)
		executed++
		if err != nil {
			(*failures)++
			for j := i + 1; j < len(prepared); j++ {
				results = append(results, commandResult{Index: j + 1, Command: commandName(prepared[j].Call.Name), Status: "not_executed"})
			}
			if storageErr != nil || !recoverableCommandError(err) {
				stopErr = err
			} else if *failures >= 3 {
				run.StopReason = "tool_error_limit"
				stopErr = ErrAgentLimit
			}
			break
		}
		succeeded++
		*failures = 0
	}
	output := encoded(map[string]any{"ok": succeeded == len(prepared), "total": len(prepared), "executed": executed, "succeeded": succeeded, "not_executed": len(prepared) - executed, "partial": succeeded > 0 && succeeded < len(prepared), "results": results})
	if err := c.Repository.SaveArtifact(ctx, stableID("chain", run.ID+":"+provider.CallID)+"-result", output); err != nil {
		return nil, err
	}
	return output, stopErr
}
