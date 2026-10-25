package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"yafa/backend/entities"
)

func hostedCommand(name string) bool { return name == "generate_image" || name == "run_analysis" }
func usedToolSlots(run entities.AgentRun) int {
	count := len(run.Tools)
	for _, call := range run.HostedCalls {
		count += call.ToolCalls
	}
	return count
}
func (c *AgentHarness) HostedCapabilities() map[string]bool {
	return map[string]bool{
		"images.generate": c != nil && c.Features.Images && c.Files != nil && c.Hosted != nil && c.Hosted.Configured("image"),
		"analysis.run":    c != nil && c.Features.Analysis && c.Files != nil && c.Hosted != nil && c.Hosted.Configured("analysis"),
	}
}

// Called only by the active bounded loop, never by a generic execution endpoint.
func (c *AgentHarness) ExecuteHostedTool(ctx context.Context, run *entities.AgentRun, version *string, call entities.ToolCall) (json.RawMessage, error) {
	if !c.commandAllowed(call.Name) {
		return nil, cliError("command_not_allowed", "Choose an available command from commands.md.")
	}
	args, fingerprint, err := toolArguments(call)
	if err != nil {
		return nil, err
	}
	request := entities.HostedRequest{}
	switch call.Name {
	case "generate_image":
		var in struct {
			Prompt  string `json:"prompt"`
			Quality string `json:"quality"`
		}
		if strictArguments(args, &in) != nil {
			return nil, ErrInvalidRecord
		}
		request.Kind, request.Instruction, request.Quality = "image", in.Prompt, in.Quality
	case "run_analysis":
		var in struct {
			Instructions string `json:"instructions"`
		}
		if strictArguments(args, &in) != nil {
			return nil, ErrInvalidRecord
		}
		request.Kind, request.Instruction = "analysis", in.Instructions
	default:
		return nil, errors.New("tool_not_allowed")
	}
	if !validText(request.Instruction, 3000) {
		return nil, cliError("invalid_hosted_instruction", "Provide a nonempty prompt/instructions of at most 3000 UTF-8 bytes.")
	}
	if c.Hosted == nil || c.Files == nil || !c.Hosted.Configured(request.Kind) {
		return nil, &entities.ModelFailure{Code: "hosted_not_configured"}
	}
	workspace, _, err := c.Workspace.Read(ctx)
	if err != nil {
		return nil, err
	}
	if workspace.Lease == nil || workspace.Lease.RunID != run.ID || !workspace.Lease.ExpiresAt.After(time.Now()) {
		return nil, ErrAgentBusy
	}
	jobID := stableID("hosted", run.ID+":"+fingerprint)
	cached, err := c.Repository.LoadArtifact(ctx, jobID+"-result")
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return nil, err
	}
	// A pending intent after interruption is not permission to bill again.
	_, err = c.Repository.LoadArtifact(ctx, jobID+"-intent")
	if err == nil {
		return nil, &entities.ModelFailure{Code: "hosted_outcome_unconfirmed", UsageUnknown: true}
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return nil, err
	}
	if len(run.HostedCalls) >= 2 {
		run.StopReason = "hosted_call_limit"
		return nil, ErrAgentLimit
	}
	if request.Kind == "image" {
		for _, old := range run.HostedCalls {
			if old.Kind == "image" {
				run.StopReason = "image_call_limit"
				return nil, ErrAgentLimit
			}
		}
	}
	reserved := 1
	if request.Kind == "analysis" {
		reserved = 2
		remaining := run.Limits.MaxTokens - run.Usage.InputTokens - run.Usage.OutputTokens
		allowance := remaining - int64(len(request.Instruction)+4096) - 1024
		if allowance < 16 {
			run.StopReason = "token_budget"
			return nil, ErrAgentLimit
		}
		request.MaxOutputTokens = min(2048, run.Limits.MaxOutputTokens, int(allowance))
	}
	if usedToolSlots(*run)+reserved > run.Limits.MaxToolCalls {
		run.StopReason = "tool_call_limit"
		return nil, ErrAgentLimit
	}
	intent := encoded(map[string]any{"run_id": run.ID, "kind": request.Kind, "instruction": request.Instruction, "max_output_tokens": request.MaxOutputTokens, "max_hosted_tool_calls": reserved})
	if err = c.Repository.SaveArtifact(ctx, jobID+"-intent", intent); err != nil {
		return nil, err
	}
	run.HostedCalls = append(run.HostedCalls, entities.HostedCall{ID: jobID, Kind: request.Kind, Status: "requested"})
	at := len(run.HostedCalls) - 1
	// Hosted fees are not covered by the configured main-model text prices.
	priorUnpriced := run.Usage.AdditionalCostsUnpriced
	run.Usage.AdditionalCostsUnpriced = true
	run.Usage.EstimatedCostUSD = nil
	if err = c.persistRun(ctx, version, *run); err != nil {
		return nil, err
	}
	budgetDay := ""
	if c.Safety != nil {
		budgetDay, err = c.Safety.Reserve(ctx, jobID, request.Kind, 0, 0)
		if err != nil {
			return nil, err
		}
	}
	result, providerErr := c.Hosted.Execute(ctx, request)
	entry := &run.HostedCalls[at]
	entry.ProviderID, entry.Model, entry.Usage, entry.ProviderUsage = result.ID, result.Model, result.Usage, result.ProviderUsage
	entry.ToolCalls, entry.ContainerIDs = result.ToolCalls, result.ContainerIDs
	entry.AdditionalCostUSD = result.AdditionalCostUSD
	priced := result.AdditionalCostUSD != nil && (request.Kind != "analysis" || result.Usage.EstimatedCostUSD != nil)
	entry.Usage.AdditionalCostsUnpriced = !priced
	run.Usage.AdditionalCostsUnpriced = priorUnpriced || !priced
	if result.AdditionalCostUSD != nil {
		run.Usage.KnownCostUSD += *result.AdditionalCostUSD
	}
	if request.Kind == "analysis" {
		run.Usage.InputTokens += result.Usage.InputTokens
		run.Usage.OutputTokens += result.Usage.OutputTokens
		run.Usage.CachedTokens += result.Usage.CachedTokens
		run.Usage.ReasoningTokens += result.Usage.ReasoningTokens
		if result.Usage.EstimatedCostUSD != nil {
			run.Usage.KnownCostUSD += *result.Usage.EstimatedCostUSD
		}
	}
	var failure *entities.ModelFailure
	if result.Usage.Incomplete || (errors.As(providerErr, &failure) && failure.UsageUnknown) {
		run.Usage.Incomplete = true
	}
	refreshUsageCost(&run.Usage)
	entry.Status = "failed"
	if providerErr == nil {
		entry.Files, providerErr = c.saveGeneratedFiles(ctx, *run, jobID, result.Files)
		if providerErr == nil {
			entry.Status = "succeeded"
		}
	}
	if c.Safety != nil && budgetDay != "" {
		if err := c.Safety.Settle(ctx, budgetDay, jobID, entry.Status, 0, 0, 0, providerErr == nil, true, agentErrorCode(providerErr)); err != nil {
			return nil, err
		}
	}
	if providerErr != nil {
		entry.ErrorCode = agentErrorCode(providerErr)
	}
	// Persist usage and file references even when generation succeeded but a
	// subsequent download/storage step failed. Never retry the provider blindly.
	if err = c.Repository.SaveArtifact(ctx, jobID+"-outcome", encoded(*entry)); err != nil {
		return nil, err
	}
	if err = c.persistRun(ctx, version, *run); err != nil {
		return nil, err
	}
	if providerErr != nil {
		// Once a hosted request started, storage/domain-looking failures must
		// not be mistaken for a safe argument correction and billed again.
		return nil, &entities.ModelFailure{Code: agentErrorCode(providerErr), UsageUnknown: run.Usage.Incomplete}
	}
	if run.Usage.Incomplete {
		return nil, &entities.ModelFailure{Code: "usage_unavailable", UsageUnknown: true}
	}
	if run.Usage.InputTokens+run.Usage.OutputTokens > run.Limits.MaxTokens || usedToolSlots(*run) > run.Limits.MaxToolCalls {
		run.StopReason = "hosted_budget_exceeded"
		return nil, ErrAgentLimit
	}
	files := []map[string]any{}
	for _, file := range entry.Files {
		files = append(files, map[string]any{"id": file.ID, "name": file.Name, "mime_type": file.MIME, "size_bytes": file.Size, "sha256": file.SHA256, "download_path": "/v1/files/" + file.ID + "/content"})
	}
	output := encoded(map[string]any{"job_id": jobID, "text": clipText(result.Text, 3000), "files": files, "files_stored": len(files), "download_requires_auth": true, "cost_status": "Additional image/container fees are unpriced; not zero. Only listed stored files are downloadable. No files are attached to this conversation unless listed here."})
	if err = c.Repository.SaveArtifact(ctx, jobID+"-result", output); err != nil {
		return nil, err
	}
	return output, nil
}
