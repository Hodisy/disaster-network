package repositories

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"yafa/backend/core"
	"yafa/backend/entities"
)

// Immutable bounded chunks share Strata's real volume/budget. The manifest is
// committed last; partial writes never become a complete downloadable file.
type FileMemory struct{ AgentMemory }

func (r FileMemory) SaveFile(ctx context.Context, meta entities.StoredFile, data []byte) error {
	if !core.ValidFileID(meta.ID) || meta.Size != len(data) || meta.Size < 1 || meta.Size > core.MaxGeneratedFileBytes || meta.SHA256 != core.FileDigest(data) || meta.Chunks != (len(data)+core.FileChunkBytes-1)/core.FileChunkBytes {
		return core.ErrInvalidRecord
	}
	for i := 0; i < meta.Chunks; i++ {
		start := i * core.FileChunkBytes
		payload, _ := json.Marshal(map[string]string{"data": base64.StdEncoding.EncodeToString(data[start:min(start+core.FileChunkBytes, len(data))])})
		if err := r.SaveArtifact(ctx, fmt.Sprintf("%s-part-%02d", meta.ID, i), payload); err != nil {
			return err
		}
	}
	payload, _ := json.Marshal(meta)
	return r.SaveArtifact(ctx, meta.ID, payload)
}
func (r FileMemory) FileMetadata(ctx context.Context, id string) (entities.StoredFile, error) {
	var meta entities.StoredFile
	if !core.ValidFileID(id) {
		return meta, core.ErrInvalidRecord
	}
	raw, err := r.LoadArtifact(ctx, id)
	if err != nil {
		return meta, err
	}
	if json.Unmarshal(raw, &meta) != nil || meta.ID != id || meta.Size < 1 || meta.Size > core.MaxGeneratedFileBytes || meta.Chunks != (meta.Size+core.FileChunkBytes-1)/core.FileChunkBytes || len(meta.SHA256) != 64 {
		return meta, core.ErrMemoryUnavailable
	}
	return meta, nil
}
func (r FileMemory) FileContent(ctx context.Context, id string) (entities.StoredFile, []byte, error) {
	meta, err := r.FileMetadata(ctx, id)
	if err != nil {
		return meta, nil, err
	}
	data := make([]byte, 0, meta.Size)
	for i := 0; i < meta.Chunks; i++ {
		raw, err := r.LoadArtifact(ctx, fmt.Sprintf("%s-part-%02d", id, i))
		if errors.Is(err, core.ErrRecordNotFound) {
			err = core.ErrMemoryUnavailable
		}
		if err != nil {
			return meta, nil, err
		}
		var part struct {
			Data string `json:"data"`
		}
		if json.Unmarshal(raw, &part) != nil || len(part.Data) > 4*((core.FileChunkBytes+2)/3) {
			return meta, nil, core.ErrMemoryUnavailable
		}
		block, err := base64.StdEncoding.DecodeString(part.Data)
		expected := min(core.FileChunkBytes, meta.Size-len(data))
		if err != nil || len(block) != expected {
			return meta, nil, core.ErrMemoryUnavailable
		}
		data = append(data, block...)
	}
	name, mime, typeErr := core.GeneratedFileType(meta.Name, data)
	if len(data) != meta.Size || core.FileDigest(data) != meta.SHA256 || typeErr != nil || name != meta.Name || mime != meta.MIME {
		return meta, nil, core.ErrMemoryUnavailable
	}
	return meta, data, nil
}
