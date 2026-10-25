package services

import (
	"context"
	"yafa/backend/entities"
)

// AudioTransport is implemented by a real transport/provider connector later.
// False means that the path is unavailable, not that silent audio is playing.
type AudioTransport interface {
	Available(ctx context.Context) (bool, error)
	SupportsFormat(ctx context.Context, format entities.AudioFormat) (bool, error)
}
