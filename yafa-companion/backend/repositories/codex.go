package repositories

import (
	"context"
	"yafa/backend/entities"
)

type CodexRepository interface {
	LoadCodexJob(context.Context, string) (entities.CodexJob, string, error)
	SaveCodexJob(context.Context, string, entities.CodexJob) (string, error)
	LoadCodexIndex(context.Context) (entities.CodexIndex, string, error)
	SaveCodexIndex(context.Context, string, entities.CodexIndex) (string, error)
}
