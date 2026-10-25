package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Keep Europe/Paris available in the minimal runtime image.
	"yafa/backend/entities"
	"yafa/backend/prompts"
)

var ErrContextNotInitialized = errors.New("context_files_not_initialized")

// InitializeContextFiles seeds only absent documents before the API starts.
// One compare-and-write commits all missing documents. Conflicts are reread;
// ambiguous transport outcomes are never blindly retried or locally replaced.
func (c WorkspaceCore) InitializeContextFiles(ctx context.Context) error {
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		workspace, version, err := c.Read(ctx)
		if err != nil {
			return err
		}
		missing := []string{}
		for _, name := range prompts.Names() {
			if _, exists := workspace.ContextFiles[name]; exists {
				// Existing content is authoritative, including edited templates.
				// Reject invalid records without replacing or repairing them.
				if _, err := contextFile(workspace, name); err != nil {
					return err
				}
			} else {
				missing = append(missing, name)
			}
		}
		if len(missing) == 0 {
			return nil
		}
		if workspace.Lease != nil && c.Now().Before(workspace.Lease.ExpiresAt) {
			return ErrAgentBusy
		}
		if workspace.ContextFiles == nil {
			workspace.ContextFiles = map[string]entities.ContextFile{}
		}
		for _, name := range missing {
			content, err := prompts.Read(name)
			if err != nil || !prompts.ValidDocument(content) {
				return errors.New("prompt_configuration_invalid")
			}
			workspace.ContextFiles[name] = entities.ContextFile{Name: name, Content: content, Revision: "1", Source: "strata"}
		}
		_, err = c.Repository.SaveWorkspace(ctx, version, workspace)
		if !errors.Is(err, ErrVersionConflict) {
			return err
		}
	}
	return ErrVersionConflict
}

// Context files are versioned workspace configuration, read only from Strata.
// Embedded files are startup seeds, never a fallback for missing runtime data.
func contextFile(workspace entities.Workspace, name string) (entities.ContextFile, error) {
	if !prompts.Editable(name) {
		return entities.ContextFile{}, ErrInvalidRecord
	}
	if file, ok := workspace.ContextFiles[name]; ok {
		if file.Name != name || !prompts.ValidDocument(file.Content) || file.Source != "strata" {
			return entities.ContextFile{}, ErrMemoryUnavailable
		}
		if _, err := revision(file.Revision); err != nil {
			return entities.ContextFile{}, ErrMemoryUnavailable
		}
		return file, nil
	}
	return entities.ContextFile{}, ErrContextNotInitialized
}

func (c WorkspaceCore) ReadContextFile(ctx context.Context, name string) (entities.ContextFile, error) {
	workspace, _, err := c.Read(ctx)
	if err != nil {
		return entities.ContextFile{}, err
	}
	return contextFile(workspace, name)
}

type UpdateContextFileInput struct {
	Name             string `json:"name"`
	ExpectedRevision string `json:"expected_revision"`
	Content          string `json:"content"`
}

func (c WorkspaceCore) UpdateContextFile(ctx context.Context, runID, actionID, fingerprint string, in UpdateContextFileInput) (json.RawMessage, error) {
	if !prompts.Editable(in.Name) || !prompts.ValidDocument(in.Content) {
		return nil, ErrInvalidRecord
	}
	return c.mutate(ctx, runID, actionID, fingerprint, func(workspace *entities.Workspace) (any, error) {
		old, err := contextFile(*workspace, in.Name)
		if err != nil {
			return nil, err
		}
		if old.Revision != in.ExpectedRevision {
			return nil, ErrVersionConflict
		}
		next, err := revision(old.Revision)
		if err != nil {
			return nil, err
		}
		if workspace.ContextFiles == nil {
			workspace.ContextFiles = map[string]entities.ContextFile{}
		}
		file := entities.ContextFile{Name: in.Name, Content: in.Content, Revision: next, Source: "strata"}
		workspace.ContextFiles[in.Name] = file
		return file, nil
	})
}

// BuildPrompt freezes one snapshot per run/session. Tool results and messages are
// sent separately; values are substituted once, without evaluating user content.
func (c *AgentHarness) BuildPrompt(ctx context.Context, thread entities.Thread, mode string, now time.Time, voiceModel string) (entities.PromptSnapshot, error) {
	if mode != "chat" && mode != "live" {
		return entities.PromptSnapshot{}, ErrInvalidRecord
	}
	workspace, _, err := c.Workspace.Read(ctx)
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	zone, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	local := now.In(zone)
	start, err := agendaDate(local.Format("2006-01-02"))
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	today, err := c.agenda(ctx, workspace, AgendaInput{Owner: "both", IncludeCompleted: true, Offset: 0, Limit: 12}, start, 1)
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	dailyState, err := c.ReadDailyState(ctx, nil, now)
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	values := map[string]string{
		"today": local.Format("Monday, 2006-01-02"), "now": local.Format(time.RFC3339), "timezone": "Europe/Paris",
		"thread_id": thread.ID, "origin": thread.Origin, "mode": mode,
		"backend_model": c.Model.Name(), "voice_model": voiceModel, "tools": commandCatalog(),
		"project_count": strconv.Itoa(len(workspace.Projects)), "task_count": strconv.Itoa(len(workspace.Tasks)),
		"other_conversations": "Today includes pointers to actually indexed conversations. Use agenda.day/week/month/range for dated activity and conversations.read for a known thread. The index does not backfill older history; no semantic/full-history search is registered.",
		"today_agenda":        string(encoded(compactAgenda(today, workspace))),
		"personal_context":    personalOrientation(workspace, local.Format("2006-01-02")),
		"today_state":         string(encoded(map[string]any{"personal_weather": dailyState, "work_progress": workProgress(today, workspace)})),
		"runtime_state":       "Projects, tasks, context and execution records use Strata. Only commands in the injected help are available. A due date does not schedule a job. Provider configuration does not prove account access.",
	}
	values["commands"], err = c.CommandHelp()
	if err != nil {
		return entities.PromptSnapshot{}, err
	}
	if voiceModel == "" {
		values["voice_model"] = "Not used by this text run."
	}
	snapshot := entities.PromptSnapshot{CreatedAt: now.UTC(), Documents: []entities.PromptDocumentRef{}}
	// Expand source documents against runtime values only, never against each other.
	blocks := map[string]string{}
	for _, name := range prompts.Names() {
		file, err := contextFile(workspace, name)
		if err != nil {
			return snapshot, err
		}
		content, err := prompts.Render(file.Content, values)
		if err != nil {
			return snapshot, errors.New("prompt_configuration_invalid")
		}
		blocks[strings.TrimSuffix(name, ".md")] = content
		snapshot.Documents = append(snapshot.Documents, entities.PromptDocumentRef{Name: name, Revision: file.Revision, Source: file.Source, Hash: stableID("document", file.Content)})
	}
	for key, content := range blocks {
		values[key] = content
	}
	channel, err := prompts.Read(mode + ".md")
	if err != nil {
		return snapshot, err
	}
	values["channel"] = channel
	values["optional_tool_notes"] = c.optionalHelp()
	template, err := prompts.Read("prompt.md")
	if err != nil {
		return snapshot, err
	}
	snapshot.Instructions, err = prompts.Render(template, values)
	if err != nil {
		return snapshot, err
	}
	if len(snapshot.Instructions) > 32768 {
		return snapshot, ErrAgentCapacity
	}
	snapshot.Hash = stableID("prompt", snapshot.Instructions)
	return snapshot, nil
}

func (c *AgentHarness) savePrompt(ctx context.Context, id string, snapshot entities.PromptSnapshot) error {
	if err := c.Repository.SaveArtifact(ctx, id+"-prompt", encoded(snapshot)); err != nil {
		return fmt.Errorf("save prompt snapshot: %w", err)
	}
	return nil
}
