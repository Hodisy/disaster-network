package services

import (
	"context"
	"yafa/backend/entities"
)

type CodexRunner interface {
	Configured() bool
	Run(context.Context, entities.CodexRequest) (entities.CodexResult, error)
}
