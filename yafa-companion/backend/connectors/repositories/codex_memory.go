package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"yafa/backend/core"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

type CodexMemory struct{ Memory repositories.MemoryRepository }

func (c CodexMemory) LoadCodexJob(ctx context.Context, id string) (entities.CodexJob, string, error) {
	var job entities.CodexJob
	if !core.ValidCodexID(id) {
		return job, "", core.ErrInvalidRecord
	}
	row, err := c.Memory.Read(ctx, id)
	if err != nil {
		return job, "", err
	}
	if json.Unmarshal(row.Value, &job) != nil || job.ID != id || job.CreatedAt.IsZero() || job.Deadline.IsZero() {
		return job, "", core.ErrMemoryUnavailable
	}
	return job, row.Version, nil
}
func (c CodexMemory) SaveCodexJob(ctx context.Context, version string, job entities.CodexJob) (string, error) {
	if !core.ValidCodexID(job.ID) {
		return "", core.ErrInvalidRecord
	}
	data, err := json.Marshal(job)
	if err != nil || len(data) > 131072 {
		return "", core.ErrInvalidRecord
	}
	row, err := c.Memory.Write(ctx, job.ID, version, data)
	return row.Version, err
}
func (c CodexMemory) LoadCodexIndex(ctx context.Context) (entities.CodexIndex, string, error) {
	value := entities.CodexIndex{IDs: []string{}}
	row, err := c.Memory.Read(ctx, "codex-jobs-v1")
	if errors.Is(err, core.ErrRecordNotFound) {
		return value, "0", nil
	}
	if err != nil {
		return value, "", err
	}
	if json.Unmarshal(row.Value, &value) != nil || len(value.IDs) > 128 {
		return value, "", core.ErrMemoryUnavailable
	}
	return value, row.Version, nil
}
func (c CodexMemory) SaveCodexIndex(ctx context.Context, version string, index entities.CodexIndex) (string, error) {
	if len(index.IDs) > 128 {
		return "", core.ErrAgentCapacity
	}
	data, err := json.Marshal(index)
	if err != nil {
		return "", err
	}
	row, err := c.Memory.Write(ctx, "codex-jobs-v1", version, data)
	return row.Version, err
}
