package repositories

import (
	"context"
	"encoding/json"
	"yafa/backend/core"
	"yafa/backend/entities"
	ports "yafa/backend/repositories"
)

type DeviceLinkMemory struct{ Memory ports.MemoryRepository }

func (r DeviceLinkMemory) ReadLink(ctx context.Context, id string) (entities.DeviceLink, string, error) {
	if r.Memory == nil {
		return entities.DeviceLink{}, "", core.ErrMemoryUnavailable
	}
	record, err := r.Memory.Read(ctx, "device."+id)
	if err != nil {
		return entities.DeviceLink{}, "", err
	}
	var link entities.DeviceLink
	if json.Unmarshal(record.Value, &link) != nil || link.DeviceID != id {
		return link, "", core.ErrInvalidRecord
	}
	return link, record.Version, nil
}

func (r DeviceLinkMemory) WriteLink(ctx context.Context, id, version string, link entities.DeviceLink) error {
	if r.Memory == nil {
		return core.ErrMemoryUnavailable
	}
	value, err := json.Marshal(link)
	if err != nil {
		return core.ErrInvalidRecord
	}
	_, err = r.Memory.Write(ctx, "device."+id, version, value)
	return err
}
