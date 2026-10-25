package repositories

import (
	"context"
	"yafa/backend/entities"
)

type FileRepository interface {
	SaveFile(context.Context, entities.StoredFile, []byte) error
	FileMetadata(context.Context, string) (entities.StoredFile, error)
	FileContent(context.Context, string) (entities.StoredFile, []byte, error)
}
