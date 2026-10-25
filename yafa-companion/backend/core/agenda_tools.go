package core

import (
	"context"
	"encoding/json"
	"time"
	"yafa/backend/entities"
)

func readOnlyAgentTool(name string) bool {
	return name == "list_memory" || name == "read_memory" || name == "read_codex" || name == "list_projects" || name == "list_tasks" || name == "read_context_file" || name == "read_daily_state" || name == "get_day" || name == "get_week" || name == "get_month" || name == "get_range" || name == "read_conversation"
}

func (c *AgentHarness) ExecuteTool(ctx context.Context, runID, threadID string, call entities.ToolCall) (json.RawMessage, string, error) {
	if !c.commandAllowed(call.Name) {
		return nil, "", cliError("command_not_allowed", "Choose an available command from commands.md.")
	}
	if codexCommand(call.Name) {
		value, err := c.executeCodex(ctx, runID, threadID, call)
		return value, "", err
	}
	if call.Name == "read_daily_state" || call.Name == "record_daily_state" {
		args, fingerprint, err := toolArguments(call)
		if err != nil {
			return nil, "", err
		}
		if call.Name == "read_daily_state" {
			var in struct {
				Date *string `json:"date"`
			}
			if err = strictArguments(args, &in); err != nil {
				return nil, fingerprint, err
			}
			value, err := c.ReadDailyState(ctx, in.Date, time.Now())
			if err != nil {
				return nil, fingerprint, err
			}
			return encoded(value), fingerprint, nil
		}
		var in RecordDailyStateInput
		if err = strictArguments(args, &in); err != nil {
			return nil, fingerprint, err
		}
		value, err := c.RecordDailyState(ctx, runID, threadID, in, fingerprint)
		return value, fingerprint, err
	}
	if call.Name != "get_day" && call.Name != "get_week" && call.Name != "get_month" && call.Name != "get_range" && call.Name != "read_conversation" {
		return c.Workspace.ExecuteTool(ctx, runID, threadID, call)
	}
	args, fingerprint, err := toolArguments(call)
	if err != nil {
		return nil, "", err
	}
	var value any
	if call.Name == "read_conversation" {
		var in struct {
			ThreadID string `json:"thread_id"`
		}
		if err = strictArguments(args, &in); err == nil {
			value, err = c.ReadConversation(ctx, in.ThreadID)
		}
	} else {
		var in struct {
			Date             *string `json:"date"`
			WeekOf           *string `json:"week_of"`
			Month            *string `json:"month"`
			StartDate        string  `json:"start_date"`
			EndDate          string  `json:"end_date"`
			Owner            string  `json:"owner"`
			IncludeCompleted bool    `json:"include_completed"`
			Offset           int     `json:"offset"`
			Limit            int     `json:"limit"`
		}
		if err = strictArguments(args, &in); err == nil {
			date := in.Date
			if call.Name == "get_week" {
				date = in.WeekOf
			}
			options := AgendaInput{Date: date, Owner: in.Owner, IncludeCompleted: in.IncludeCompleted, Offset: in.Offset, Limit: in.Limit}
			switch call.Name {
			case "get_month":
				value, err = c.ReadAgendaMonth(ctx, options, in.Month, time.Now())
			case "get_range":
				value, err = c.ReadAgendaRange(ctx, options, in.StartDate, in.EndDate)
			default:
				value, err = c.ReadAgenda(ctx, options, call.Name == "get_week", time.Now())
			}
		}
	}
	if err != nil {
		return nil, fingerprint, err
	}
	data, err := json.Marshal(value)
	return data, fingerprint, err
}

func (c *AgentHarness) ReadConversation(ctx context.Context, id string) (any, error) {
	thread, err := c.Thread(ctx, id)
	if err != nil {
		return nil, err
	}
	from := max(0, len(thread.Messages)-8)
	messages := []map[string]any{}
	for _, message := range thread.Messages[from:] {
		messages = append(messages, map[string]any{"id": message.ID, "role": message.Role, "text": clipText(message.Text, 1000), "text_truncated": len(message.Text) > 1000, "created_at": message.CreatedAt, "run_id": message.RunID})
	}
	var latest any
	if len(thread.Messages) > 0 {
		run, err := c.ReadRun(ctx, thread.Messages[len(thread.Messages)-1].RunID)
		if err != nil {
			return nil, err
		}
		latest = map[string]any{"id": run.ID, "status": run.Status, "error_code": run.ErrorCode, "effects_may_have_committed": run.EffectsMayHaveCommitted}
	}
	return map[string]any{"thread_id": thread.ID, "origin": thread.Origin, "messages": messages, "earlier_messages_omitted": from, "latest_run": latest}, nil
}
