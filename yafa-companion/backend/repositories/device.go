package repositories

import (
	"context"
	"yafa/backend/entities"
)

// Implementations must read actual stored devices; no fabricated fallback records.
type DeviceRepository interface {
	GetDevice(ctx context.Context, deviceID string) (entities.Device, error)
}
