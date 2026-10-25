package repositories

import (
	"context"
	"yafa/backend/entities"
)

type DeviceLinkRepository interface {
	ReadLink(context.Context, string) (entities.DeviceLink, string, error)
	WriteLink(context.Context, string, string, entities.DeviceLink) error
}
