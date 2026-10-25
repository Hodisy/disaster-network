package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"yafa/backend/core"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

type AgentMemory struct{ Memory repositories.MemoryRepository }

func (r AgentMemory) read(ctx context.Context, key string, out any) (string, error) {
	row, err := r.Memory.Read(ctx, key)
	if err != nil {
		return "", err
	}
	if json.Unmarshal(row.Value, out) != nil {
		return "", core.ErrMemoryUnavailable
	}
	return row.Version, nil
}
func (r AgentMemory) save(ctx context.Context, key, version string, value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", core.ErrInvalidRecord
	}
	if len(data) > 196608 {
		return "", core.ErrAgentCapacity
	}
	row, err := r.Memory.Write(ctx, key, version, data)
	if err != nil {
		return "", err
	}
	return row.Version, nil
}
func (r AgentMemory) LoadWorkspace(ctx context.Context) (entities.Workspace, string, error) {
	var value entities.Workspace
	version, err := r.read(ctx, "agent-workspace-v1", &value)
	if err == nil && (value.Projects == nil || value.Tasks == nil || value.Actions == nil) {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveWorkspace(ctx context.Context, version string, value entities.Workspace) (string, error) {
	return r.save(ctx, "agent-workspace-v1", version, value)
}
func (r AgentMemory) LoadThread(ctx context.Context, id string) (entities.Thread, string, error) {
	var value entities.Thread
	version, err := r.read(ctx, id, &value)
	if err == nil && value.ID != id {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveThread(ctx context.Context, version string, value entities.Thread) (string, error) {
	return r.save(ctx, value.ID, version, value)
}
func (r AgentMemory) LoadRun(ctx context.Context, id string) (entities.AgentRun, string, error) {
	var value entities.AgentRun
	version, err := r.read(ctx, id, &value)
	if err == nil && value.ID != id {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveRun(ctx context.Context, version string, value entities.AgentRun) (string, error) {
	return r.save(ctx, value.ID, version, value)
}
func (r AgentMemory) SaveArtifact(ctx context.Context, key string, data json.RawMessage) error {
	if len(data) > 196608 || !json.Valid(data) {
		return core.ErrAgentCapacity
	}
	_, err := r.Memory.Write(ctx, key, "0", data)
	if errors.Is(err, core.ErrVersionConflict) {
		// Immutable artifact retries must match the exact already-committed content.
		row, readErr := r.Memory.Read(ctx, key)
		if readErr != nil {
			return readErr
		}
		var left, right any
		if json.Unmarshal(row.Value, &left) != nil || json.Unmarshal(data, &right) != nil {
			return core.ErrMemoryUnavailable
		}
		a, _ := json.Marshal(left)
		b, _ := json.Marshal(right)
		if string(a) == string(b) {
			return nil
		}
	}
	return err
}
func (r AgentMemory) LoadVoice(ctx context.Context, id string) (entities.VoiceRecord, string, error) {
	var value entities.VoiceRecord
	version, err := r.read(ctx, id, &value)
	if err == nil && value.ID != id {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveVoice(ctx context.Context, version string, value entities.VoiceRecord) (string, error) {
	return r.save(ctx, value.ID, version, value)
}

func (r AgentMemory) LoadDayActivity(ctx context.Context, date string) (entities.DayActivity, string, error) {
	var value entities.DayActivity
	version, err := r.read(ctx, "agent-day-"+date, &value)
	if err == nil && (value.Date != date || value.Threads == nil) {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveDayActivity(ctx context.Context, version string, value entities.DayActivity) (string, error) {
	return r.save(ctx, "agent-day-"+value.Date, version, value)
}

func (r AgentMemory) LoadDailyState(ctx context.Context, date string) (entities.DailyState, string, error) {
	var value entities.DailyState
	version, err := r.read(ctx, "personal-state-"+date, &value)
	if err == nil && (value.Date != date || value.Gauges == nil || value.Actions == nil) {
		err = core.ErrMemoryUnavailable
	}
	return value, version, err
}
func (r AgentMemory) SaveDailyState(ctx context.Context, version string, value entities.DailyState) (string, error) {
	return r.save(ctx, "personal-state-"+value.Date, version, value)
}

func (r AgentMemory) LoadArtifact(ctx context.Context, key string) (json.RawMessage, error) {
	row, err := r.Memory.Read(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(row.Value) > 196608 || !json.Valid(row.Value) {
		return nil, core.ErrMemoryUnavailable
	}
	return row.Value, nil
}
