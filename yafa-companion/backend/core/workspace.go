package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

var (
	ErrAgentBusy        = errors.New("agent_busy")
	ErrAgentCapacity    = errors.New("agent_capacity_reached")
	ErrAgentUnavailable = errors.New("agent_not_configured")
	ErrAgentLimit       = errors.New("agent_limit_reached")
)
var agentEventPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}$`)
var agentIDPattern = regexp.MustCompile(`^(thread|run|project|task)-[a-f0-9]{32}$`)

func validAgentID(id, prefix string) bool {
	return agentIDPattern.MatchString(id) && strings.HasPrefix(id, prefix+"-")
}

func stableID(prefix, seed string) string {
	hash := sha256.Sum256([]byte(seed))
	return prefix + "-" + hex.EncodeToString(hash[:16])
}
func validText(s string, max int) bool {
	return strings.TrimSpace(s) != "" && len(s) <= max && utf8.ValidString(s)
}
func strictArguments(data string, out any) error {
	if len(data) > 8192 {
		return ErrInvalidRecord
	}
	d := json.NewDecoder(strings.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(out) != nil || d.Decode(new(any)) != io.EOF {
		return ErrInvalidRecord
	}
	return nil
}
func revision(value string) (string, error) {
	n, err := strconv.ParseUint(value, 10, 64)
	if err != nil || n == 0 || n == math.MaxUint64 {
		return "", ErrInvalidRecord
	}
	return strconv.FormatUint(n+1, 10), nil
}

type WorkspaceCore struct {
	Repository repositories.AgentRepository
	Now        func() time.Time
}

func (c WorkspaceCore) Read(ctx context.Context) (entities.Workspace, string, error) {
	value, version, err := c.Repository.LoadWorkspace(ctx)
	if errors.Is(err, ErrRecordNotFound) {
		// A new installation is empty, never a fabricated fallback for storage failure.
		return entities.Workspace{Projects: map[string]entities.Project{}, Tasks: map[string]entities.Task{}, Actions: map[string]entities.ActionReceipt{}}, "0", nil
	}
	return value, version, err
}
func (c WorkspaceCore) Lease(ctx context.Context, runID string, until *time.Time) error {
	for attempt := 0; attempt < 3; attempt++ {
		value, version, err := c.Read(ctx)
		if err != nil {
			return err
		}
		if value.Lease != nil && value.Lease.RunID != runID && c.Now().Before(value.Lease.ExpiresAt) {
			return ErrAgentBusy
		}
		if until == nil {
			if value.Lease == nil || value.Lease.RunID != runID {
				return nil
			}
			value.Lease = nil
		} else {
			value.Lease = &entities.RunLease{RunID: runID, ExpiresAt: *until}
		}
		_, err = c.Repository.SaveWorkspace(ctx, version, value)
		if !errors.Is(err, ErrVersionConflict) {
			return err
		}
	}
	return ErrVersionConflict
}
func (c WorkspaceCore) mutate(ctx context.Context, runID, actionID, fingerprint string, apply func(*entities.Workspace) (any, error)) (json.RawMessage, error) {
	if !validText(actionID, 96) || !validText(fingerprint, 96) {
		return nil, ErrInvalidRecord
	}
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		value, version, err := c.Read(ctx)
		if err != nil {
			return nil, err
		}
		if receipt, exists := value.Actions[actionID]; exists {
			if receipt.Fingerprint != fingerprint {
				return nil, ErrVersionConflict
			}
			return receipt.Result, nil
		}
		if runID != "" {
			if value.Lease == nil || value.Lease.RunID != runID || !c.Now().Before(value.Lease.ExpiresAt) {
				return nil, ErrAgentBusy
			}
		} else if value.Lease != nil && c.Now().Before(value.Lease.ExpiresAt) {
			return nil, ErrAgentBusy
		}
		if len(value.Actions) >= 384 {
			return nil, ErrAgentCapacity
		}
		result, err := apply(&value)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(result)
		if err != nil {
			return nil, ErrInvalidRecord
		}
		value.Actions[actionID] = entities.ActionReceipt{Fingerprint: fingerprint, Result: data}
		_, err = c.Repository.SaveWorkspace(ctx, version, value)
		if errors.Is(err, ErrVersionConflict) {
			continue
		}
		// Transport errors are not retried: a write may already have committed.
		return data, err
	}
	return nil, ErrVersionConflict
}
func (c WorkspaceCore) Projects(ctx context.Context, archived bool, offset, limit int) (any, error) {
	if offset < 0 || offset > 64 || limit < 1 || limit > 20 {
		return nil, ErrInvalidRecord
	}
	w, _, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	items := []entities.Project{}
	for _, item := range w.Projects {
		if archived || !item.Archived {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	start := min(offset, len(items))
	end := min(start+limit, len(items))
	var next *int
	if end < len(items) {
		next = &end
	}
	return map[string]any{"projects": items[start:end], "next_offset": next}, nil
}
func (c WorkspaceCore) Tasks(ctx context.Context, projectID *string, completed bool, offset, limit int) (any, error) {
	if offset < 0 || offset > 128 || limit < 1 || limit > 20 {
		return nil, ErrInvalidRecord
	}
	w, _, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	items := []entities.Task{}
	for _, item := range w.Tasks {
		if !completed && item.Completed {
			continue
		}
		if projectID != nil && (item.ProjectID == nil || *item.ProjectID != *projectID) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	start := min(offset, len(items))
	end := min(start+limit, len(items))
	var next *int
	if end < len(items) {
		next = &end
	}
	return map[string]any{"tasks": items[start:end], "next_offset": next}, nil
}

type CreateProjectInput struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

func (c WorkspaceCore) CreateProject(ctx context.Context, runID, actionID, fingerprint string, in CreateProjectInput) (json.RawMessage, error) {
	if !validText(in.Name, 120) || len(in.Description) > 1200 || (in.Kind != "personal" && in.Kind != "professional") {
		return nil, ErrInvalidRecord
	}
	return c.mutate(ctx, runID, actionID, fingerprint, func(w *entities.Workspace) (any, error) {
		for _, item := range w.Projects {
			if strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(in.Name)) {
				if item.Archived || item.Kind != in.Kind {
					return nil, ErrVersionConflict
				}
				return map[string]any{"project": item, "created": false}, nil
			}
		}
		if len(w.Projects) >= 64 {
			return nil, ErrAgentCapacity
		}
		item := entities.Project{ID: stableID("project", actionID), Name: strings.TrimSpace(in.Name), Kind: in.Kind, Description: in.Description, Revision: "1"}
		w.Projects[item.ID] = item
		return map[string]any{"project": item, "created": true}, nil
	})
}

type UpdateProjectInput struct {
	ID       string  `json:"project_id"`
	Expected string  `json:"expected_revision"`
	Name     *string `json:"name"`
	Archived *bool   `json:"archived"`
}

func (c WorkspaceCore) UpdateProject(ctx context.Context, runID, actionID, fingerprint string, in UpdateProjectInput) (json.RawMessage, error) {
	if !agentIDPattern.MatchString(in.ID) || (in.Name == nil && in.Archived == nil) || (in.Name != nil && !validText(*in.Name, 120)) {
		return nil, ErrInvalidRecord
	}
	return c.mutate(ctx, runID, actionID, fingerprint, func(w *entities.Workspace) (any, error) {
		item, ok := w.Projects[in.ID]
		if !ok {
			return nil, ErrRecordNotFound
		}
		if item.Revision != in.Expected {
			return nil, ErrVersionConflict
		}
		if in.Name != nil {
			for id, p := range w.Projects {
				if id != in.ID && strings.EqualFold(strings.TrimSpace(p.Name), strings.TrimSpace(*in.Name)) {
					return nil, ErrVersionConflict
				}
			}
			item.Name = strings.TrimSpace(*in.Name)
		}
		if in.Archived != nil {
			item.Archived = *in.Archived
		}
		next, err := revision(item.Revision)
		if err != nil {
			return nil, err
		}
		item.Revision = next
		w.Projects[item.ID] = item
		return map[string]any{"project": item}, nil
	})
}

type CreateTaskInput struct {
	ProjectID     *string `json:"project_id"`
	Title         string  `json:"title"`
	Notes         string  `json:"notes"`
	DueAt         *string `json:"due_at"`
	ScheduledDate *string `json:"scheduled_date"`
	ScheduledTime *string `json:"scheduled_time"`
}

func (c WorkspaceCore) CreateTask(ctx context.Context, runID, threadID, actionID, fingerprint string, in CreateTaskInput) (json.RawMessage, error) {
	if !validText(in.Title, 240) || len(in.Notes) > 1200 || !validSchedule(in.ScheduledDate, in.ScheduledTime) {
		return nil, ErrInvalidRecord
	}
	var due *time.Time
	if in.DueAt != nil {
		value, err := time.Parse(time.RFC3339, *in.DueAt)
		if err != nil {
			return nil, ErrInvalidRecord
		}
		utc := value.UTC()
		due = &utc
	}
	return c.mutate(ctx, runID, actionID, fingerprint, func(w *entities.Workspace) (any, error) {
		if len(w.Tasks) >= 128 {
			return nil, ErrAgentCapacity
		}
		if in.ProjectID != nil {
			project, ok := w.Projects[*in.ProjectID]
			if !ok {
				return nil, ErrRecordNotFound
			}
			if project.Archived {
				return nil, ErrVersionConflict
			}
		}
		item := entities.Task{ID: stableID("task", actionID), ProjectID: in.ProjectID, Title: strings.TrimSpace(in.Title), Notes: in.Notes, DueAt: due, ScheduledDate: in.ScheduledDate, ScheduledTime: in.ScheduledTime, Revision: "1", SourceThreadID: threadID}
		w.Tasks[item.ID] = item
		return map[string]any{"task": item}, nil
	})
}

type CompleteTaskInput struct {
	ID        string `json:"task_id"`
	Expected  string `json:"expected_revision"`
	Completed bool   `json:"completed"`
}

func (c WorkspaceCore) CompleteTask(ctx context.Context, runID, actionID, fingerprint string, in CompleteTaskInput) (json.RawMessage, error) {
	return c.mutate(ctx, runID, actionID, fingerprint, func(w *entities.Workspace) (any, error) {
		item, ok := w.Tasks[in.ID]
		if !ok {
			return nil, ErrRecordNotFound
		}
		if item.Revision != in.Expected {
			return nil, ErrVersionConflict
		}
		next, err := revision(item.Revision)
		if err != nil {
			return nil, err
		}
		item.Revision = next
		item.Completed = in.Completed
		w.Tasks[item.ID] = item
		return map[string]any{"task": item}, nil
	})
}
