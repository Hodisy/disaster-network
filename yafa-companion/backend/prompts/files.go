// Package prompts owns startup seed Markdown and bounded variable expansion.
// Runtime context documents are loaded from Strata by cores, never by this package.
package prompts

import (
	"embed"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

//go:embed *.md
var files embed.FS

var variables = regexp.MustCompile(`\{\{([^{}]*)\}\}`)
var variableName = regexp.MustCompile(`^[a-z_]+$`)

func Names() []string {
	return []string{"soul.md", "directive.md", "project.md", "workspace.md", "user.md"}
}

func Editable(name string) bool {
	for _, candidate := range Names() {
		if candidate == name {
			return true
		}
	}
	return false
}

func Read(name string) (string, error) {
	if !Editable(name) && name != "prompt.md" && name != "chat.md" && name != "live.md" && name != "commands.md" {
		return "", errors.New("prompt_file_not_allowed")
	}
	data, err := files.ReadFile(name)
	return strings.TrimSpace(string(data)), err
}

// Render is single-pass: values containing {{...}} are never evaluated again.
// This is text substitution, not a template language with code or file access.
func Render(source string, values map[string]string) (string, error) {
	remaining := variables.ReplaceAllString(source, "")
	if strings.Contains(remaining, "{{") || strings.Contains(remaining, "}}") {
		return "", errors.New("invalid_prompt_variable")
	}
	missing := false
	result := variables.ReplaceAllStringFunc(source, func(match string) string {
		key := strings.TrimSpace(variables.FindStringSubmatch(match)[1])
		value, ok := values[key]
		if !ok || !variableName.MatchString(key) {
			missing = true
		}
		return value
	})
	if missing {
		return "", errors.New("unknown_prompt_variable")
	}
	return result, nil
}

func ValidDocument(content string) bool {
	if strings.TrimSpace(content) == "" || len(content) > 4096 || !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
		return false
	}
	_, err := Render(content, map[string]string{
		"today": "", "now": "", "timezone": "", "thread_id": "", "origin": "", "mode": "",
		"backend_model": "", "voice_model": "", "tools": "", "other_conversations": "",
		"project_count": "", "task_count": "", "runtime_state": "",
		"today_agenda":     "",
		"today_state":      "",
		"personal_context": "",
	})
	return err == nil
}
