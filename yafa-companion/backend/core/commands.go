package core

import (
	"sort"
	"strings"
	"yafa/backend/entities"
)

func commandName(name string) string {
	switch name {
	case "run_codex":
		return "codex.run"
	case "read_codex":
		return "codex.status"
	case "cancel_codex":
		return "codex.cancel"
	case "generate_image":
		return "images.generate"
	case "run_analysis":
		return "analysis.run"
	case "get_day":
		return "agenda.day"
	case "get_week":
		return "agenda.week"
	case "get_month":
		return "agenda.month"
	case "get_range":
		return "agenda.range"
	case "list_memory":
		return "memory.list"
	case "read_memory":
		return "memory.read"
	case "list_projects":
		return "projects.list"
	case "create_project":
		return "projects.create"
	case "update_project":
		return "projects.update"
	case "list_tasks":
		return "tasks.list"
	case "create_task":
		return "tasks.create"
	case "complete_task":
		return "tasks.complete"
	case "read_conversation":
		return "conversations.read"
	case "read_context_file":
		return "context.read"
	case "update_context_file":
		return "context.update"
	case "read_daily_state":
		return "state.read"
	case "record_daily_state":
		return "state.record"
	}
	return ""
}

// Only the small string envelope is exposed to the provider. Domain schemas
// remain internal; the compact help is generated from that same registry.
func AgentTools() []entities.ModelTool {
	return []entities.ModelTool{tool("command", "Run Yafa CLI commands from commands.md. Chain up to four with &&; stop on failure.", map[string]any{
		"command": map[string]any{"type": "string"},
	})}
}

func commandCatalog() string { return "command({command: string}) — see commands.md." }

func registeredCommands(specs []entities.ModelTool) []string {
	names := []string{}
	for _, spec := range specs {
		names = append(names, commandName(spec.Name))
	}
	return names
}

func commandReference(specs []entities.ModelTool) string {
	summaries := map[string]string{
		"list_memory": "List personal documents", "read_memory": "Read bounded personal context",
		"run_codex": "Start remote job", "read_codex": "Read remote job", "cancel_codex": "Stop remote job",
		"generate_image": "Generate image", "run_analysis": "Run hosted Python",
		"list_projects": "List projects", "create_project": "Create project", "update_project": "Edit project",
		"list_tasks": "List human tasks", "get_day": "Read day", "get_week": "Read week", "get_month": "Read month", "get_range": "Read date range",
		"read_conversation": "Read recent messages", "create_task": "Create human task", "complete_task": "Complete/reopen task",
		"read_context_file": "Read context", "update_context_file": "Edit context", "read_daily_state": "Read daily state", "record_daily_state": "Record quoted estimate",
	}
	lines := []string{}
	for _, spec := range specs {
		keys := []string{}
		properties := spec.Parameters["properties"].(map[string]any)
		for key := range properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		flags := []string{}
		for _, key := range keys {
			property := properties[key].(map[string]any)
			kind, nullable := cliPropertyType(property)
			label := map[string]string{"string": "TEXT", "integer": "INT", "boolean": "BOOL"}[kind]
			choices := []string{}
			switch values := property["enum"].(type) {
			case []string:
				choices = values
			case []any:
				for _, v := range values {
					if text, ok := v.(string); ok {
						choices = append(choices, text)
					}
				}
			}
			if len(choices) > 0 {
				label = strings.Join(choices, "|")
			}
			if min, ok := property["minimum"]; ok {
				label = string(encoded(min)) + ".." + string(encoded(property["maximum"]))
			}
			flag := "--" + strings.ReplaceAll(key, "_", "-") + " " + label
			if value, ok := cliDefault(spec.Name, key, kind, nullable); ok {
				if value == nil {
					flag += "?"
				} else {
					flag += "=" + string(encoded(value))
				}
			} else {
				flag += "!"
			}
			flags = append(flags, flag)
		}
		lines = append(lines, commandName(spec.Name)+" — "+summaries[spec.Name]+".\n  "+strings.Join(flags, " "))
	}
	return strings.Join(lines, "\n")
}

func dailyScore(description string) map[string]any {
	return map[string]any{"type": []string{"integer", "null"}, "minimum": 1, "maximum": 5, "description": description}
}
