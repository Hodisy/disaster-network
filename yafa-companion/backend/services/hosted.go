package services

import (
	"context"
	"yafa/backend/entities"
)

// Provider-owned image generation and Python sandbox; no local code execution.
type HostedTools interface {
	Configured(kind string) bool
	Execute(context.Context, entities.HostedRequest) (entities.HostedResult, error)
}
