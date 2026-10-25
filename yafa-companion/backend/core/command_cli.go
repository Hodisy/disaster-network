package core

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"yafa/backend/entities"
)

type commandError struct{ Code, Hint string }

func (e *commandError) Error() string  { return e.Code }
func (e *commandError) Unwrap() error  { return ErrInvalidRecord }
func cliError(code, hint string) error { return &commandError{Code: code, Hint: hint} }

type cliToken struct {
	Text   string
	Quoted bool
}
type preparedCommand struct {
	Call        entities.ToolCall
	Fingerprint string
}

// Split only unquoted &&. Preserve quoting/escapes for the argv lexer.
// This language never invokes a shell or expands variables and paths.
func splitCommandChain(line string) ([]string, error) {
	if !validText(line, 4096) || strings.ContainsAny(line, "\x00\n\r") {
		return nil, cliError("invalid_command_line", "Use one nonempty line, at most 4096 UTF-8 bytes. Encode text newlines as \\n.")
	}
	var quote rune
	escape := false
	start := 0
	parts := []string{}
	for i := 0; i < len(line); i++ {
		r := rune(line[i])
		if escape {
			escape = false
			continue
		}
		if r == '\\' {
			escape = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == '&' {
			if i+1 >= len(line) || line[i+1] != '&' {
				return nil, cliError("unsupported_command_syntax", "Use && between commands. A single & is not supported.")
			}
			part := strings.TrimSpace(line[start:i])
			if part == "" {
				return nil, cliError("empty_command", "Put a command on both sides of &&.")
			}
			parts = append(parts, part)
			i++
			start = i + 1
		} else if strings.ContainsRune(";|<>`$", r) {
			return nil, cliError("unsupported_command_syntax", "Only && chaining is supported. Quote literal operators; no pipes, substitutions or redirects.")
		}
	}
	if escape || quote != 0 {
		return nil, cliError("unterminated_command_quote", "Close the quote or complete the backslash escape.")
	}
	tail := strings.TrimSpace(line[start:])
	if tail == "" {
		return nil, cliError("empty_command", "Put a command after &&.")
	}
	parts = append(parts, tail)
	if len(parts) > 4 {
		return nil, cliError("command_chain_limit", "Use at most four commands per call.")
	}
	return parts, nil
}

func cliTokens(line string) ([]cliToken, error) {
	tokens := []cliToken{}
	var text strings.Builder
	var quote rune
	started, quoted, escape := false, false, false
	flush := func() {
		if started {
			tokens = append(tokens, cliToken{Text: text.String(), Quoted: quoted})
			text.Reset()
			started = false
			quoted = false
		}
	}
	for _, r := range line {
		if escape {
			switch r {
			case 'n':
				text.WriteRune('\n')
			case 'r':
				text.WriteRune('\r')
			case 't':
				text.WriteRune('\t')
			default:
				text.WriteRune(r)
			}
			escape = false
			started = true
			continue
		}
		if r == '\\' {
			escape = true
			started = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				text.WriteRune(r)
			}
			started = true
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			quoted = true
			started = true
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		text.WriteRune(r)
		started = true
	}
	if escape || quote != 0 {
		return nil, cliError("unterminated_command_quote", "Close the quote or complete the backslash escape.")
	}
	flush()
	if len(tokens) == 0 || len(tokens) > 80 {
		return nil, cliError("invalid_command_line", "Use a documented command with at most 80 tokens.")
	}
	return tokens, nil
}

func cliPropertyType(property map[string]any) (string, bool) {
	if kind, ok := property["type"].(string); ok {
		return kind, false
	}
	if kinds, ok := property["type"].([]string); ok && len(kinds) == 2 {
		return kinds[0], true
	}
	return "", false
}

func cliDefault(command, key, kind string, nullable bool) (any, bool) {
	if nullable {
		return nil, true
	}
	if command == "generate_image" && key == "quality" {
		return "low", true
	}
	if key == "limit" {
		if command == "list_memory" {
			return 10, true
		}
		if command == "read_memory" {
			return 4096, true
		}
		return 20, true
	}
	if key == "offset" {
		return 0, true
	}
	if kind == "boolean" {
		if key == "completed" && command == "complete_task" {
			return true, true
		}
		return key == "include_completed" && (command == "get_day" || command == "get_week" || command == "get_month" || command == "get_range"), true
	}
	if key == "owner" {
		return "both", true
	}
	if key == "kind" && command == "create_project" {
		return "personal", true
	}
	if key == "notes" || key == "description" {
		return "", true
	}
	return nil, false
}

func validateCLIValue(key string, value any, property map[string]any) error {
	if value == nil {
		return nil
	}
	flag := "--" + strings.ReplaceAll(key, "_", "-")
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
		found := false
		for _, choice := range choices {
			if value == choice {
				found = true
			}
		}
		if !found {
			return cliError("invalid_flag_value", flag+" must be "+strings.Join(choices, ", ")+".")
		}
	}
	if n, ok := value.(int64); ok {
		if bound, ok := property["minimum"].(int); ok && n < int64(bound) {
			return cliError("invalid_flag_value", fmt.Sprintf("%s must be at least %d.", flag, bound))
		}
		if bound, ok := property["maximum"].(int); ok && n > int64(bound) {
			return cliError("invalid_flag_value", fmt.Sprintf("%s must be at most %d.", flag, bound))
		}
	}
	return nil
}

func parseCLI(line, callID string, specs []entities.ModelTool) (preparedCommand, error) {
	args, err := cliTokens(line)
	if err != nil {
		return preparedCommand{}, err
	}
	if args[0].Text == "yafa" {
		args = args[1:]
	}
	if len(args) == 0 {
		return preparedCommand{}, cliError("missing_command", "Choose a command from commands.md.")
	}
	var spec *entities.ModelTool
	for _, candidate := range specs {
		if commandName(candidate.Name) == args[0].Text {
			copy := candidate
			spec = &copy
			break
		}
	}
	if spec == nil {
		return preparedCommand{}, cliError("command_not_allowed", "Unknown command. Choose a name from commands.md.")
	}
	properties := spec.Parameters["properties"].(map[string]any)
	values := map[string]any{}
	keys := []string{}
	for key, raw := range properties {
		keys = append(keys, key)
		kind, nullable := cliPropertyType(raw.(map[string]any))
		if value, ok := cliDefault(spec.Name, key, kind, nullable); ok {
			values[key] = value
		}
	}
	sort.Strings(keys)
	seen := map[string]bool{}
	for i := 1; i < len(args); i++ {
		flag := args[i]
		if !strings.HasPrefix(flag.Text, "--") || flag.Text == "--" {
			return preparedCommand{}, cliError("unexpected_command_argument", "Use named --flags for every argument.")
		}
		parts := strings.SplitN(strings.TrimPrefix(flag.Text, "--"), "=", 2)
		key := strings.ReplaceAll(parts[0], "-", "_")
		// The documented spelling uses hyphens, never JSON's underscores.
		raw, ok := properties[key]
		if ok && strings.Contains(parts[0], "_") {
			return preparedCommand{}, cliError("unknown_command_flag", "Use --"+strings.ReplaceAll(key, "_", "-")+" (hyphens, not underscores).")
		}
		if !ok {
			allowed := []string{}
			for _, key := range keys {
				allowed = append(allowed, "--"+strings.ReplaceAll(key, "_", "-"))
			}
			return preparedCommand{}, cliError("unknown_command_flag", "Allowed flags: "+strings.Join(allowed, ", ")+".")
		}
		name := "--" + parts[0]
		if seen[key] {
			return preparedCommand{}, cliError("duplicate_command_flag", "Use "+name+" only once.")
		}
		seen[key] = true
		property := raw.(map[string]any)
		kind, nullable := cliPropertyType(property)
		var value cliToken
		if len(parts) == 2 {
			value = cliToken{Text: parts[1], Quoted: flag.Quoted}
		} else if kind == "boolean" && (i+1 == len(args) || strings.HasPrefix(args[i+1].Text, "--")) {
			values[key] = true
			continue
		} else {
			i++
			if i >= len(args) || (strings.HasPrefix(args[i].Text, "--") && !args[i].Quoted) {
				return preparedCommand{}, cliError("missing_flag_value", "Provide a value for "+name+".")
			}
			value = args[i]
		}
		if nullable && value.Text == "null" && !value.Quoted {
			values[key] = nil
			continue
		}
		switch kind {
		case "string":
			values[key] = value.Text
		case "integer":
			n, e := strconv.ParseInt(value.Text, 10, 64)
			if e != nil {
				return preparedCommand{}, cliError("invalid_flag_value", name+" must be an integer.")
			}
			values[key] = n
		case "boolean":
			if value.Text != "true" && value.Text != "false" {
				return preparedCommand{}, cliError("invalid_flag_value", name+" must be true or false.")
			}
			values[key] = value.Text == "true"
		default:
			return preparedCommand{}, cliError("invalid_flag_type", "Unsupported flag type. Consult commands.md.")
		}
		if err := validateCLIValue(key, values[key], property); err != nil {
			return preparedCommand{}, err
		}
	}
	for _, key := range keys {
		if _, ok := values[key]; !ok {
			return preparedCommand{}, cliError("missing_required_flag", "Add --"+strings.ReplaceAll(key, "_", "-")+".")
		}
	}
	call := entities.ToolCall{CallID: callID, Name: spec.Name, Arguments: string(encoded(values))}
	_, fingerprint, err := toolArguments(call)
	if err != nil {
		return preparedCommand{}, cliError("invalid_command_arguments", "Invalid arguments. Check this command's flags in commands.md.")
	}
	return preparedCommand{Call: call, Fingerprint: fingerprint}, nil
}

// All CLI syntax, primitive types, enums and declared ranges are checked before
// any action. Stateful business rules remain authoritative in the domain cores.
func parseCommandBatch(call entities.ToolCall, specs []entities.ModelTool) ([]preparedCommand, int, error) {
	if call.Name != "command" {
		return nil, 0, errors.New("tool_not_allowed")
	}
	var in struct {
		Command string `json:"command"`
	}
	if strictArguments(call.Arguments, &in) != nil {
		return nil, 0, cliError("invalid_command_envelope", "Pass exactly {command: string}.")
	}
	lines, err := splitCommandChain(in.Command)
	if err != nil {
		return nil, 0, err
	}
	prepared := []preparedCommand{}
	for i, line := range lines {
		item, err := parseCLI(line, call.CallID, specs)
		if err != nil {
			return nil, i + 1, err
		}
		prepared = append(prepared, item)
	}
	return prepared, 0, nil
}

// Validation only: no storage mutation, provider request or tool execution.
func validateCommands(line string, specs []entities.ModelTool) ([]string, error) {
	prepared, _, err := parseCommandBatch(entities.ToolCall{Name: "command", Arguments: string(encoded(map[string]string{"command": line}))}, specs)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, item := range prepared {
		names = append(names, commandName(item.Call.Name))
	}
	return names, nil
}
func CommandValidationMessage(err error) string { return commandFailureFor(err, "command").Message }

// Standalone validation defaults to built-in commands only.
func ValidateCommands(line string) ([]string, error) {
	return validateCommands(line, (&AgentHarness{}).CommandSpecs())
}
func (c *AgentHarness) ValidateCommands(line string) ([]string, error) {
	return validateCommands(line, c.CommandSpecs())
}
