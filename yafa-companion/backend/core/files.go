package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
	"yafa/backend/entities"
)

const MaxGeneratedFileBytes = 4 * 1024 * 1024
const MaxGeneratedTotalBytes = 6 * 1024 * 1024
const FileChunkBytes = 96 * 1024

var fileIDPattern = regexp.MustCompile(`^file-[a-f0-9]{32}$`)
var safeFileName = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

func ValidFileID(id string) bool { return fileIDPattern.MatchString(id) }

// Only inert download formats; MIME/signatures are checked again after storage.
func GeneratedFileType(name string, data []byte) (string, string, error) {
	if len(data) == 0 || len(data) > MaxGeneratedFileBytes {
		return "", "", ErrAgentCapacity
	}
	ext := strings.ToLower(path.Ext(name))
	mime := ""
	switch ext {
	case ".png":
		if bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
			mime = "image/png"
		}
	case ".jpg", ".jpeg":
		if bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}) {
			mime = "image/jpeg"
		}
	case ".webp":
		if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
			mime = "image/webp"
		}
	case ".pdf":
		if bytes.HasPrefix(data, []byte("%PDF-")) {
			mime = "application/pdf"
		}
	case ".txt", ".csv", ".json":
		if utf8.Valid(data) && !bytes.ContainsRune(data, 0) {
			mime = map[string]string{".txt": "text/plain", ".csv": "text/csv", ".json": "application/json"}[ext]
			if ext == ".json" && !json.Valid(data) {
				mime = ""
			}
		}
	}
	if mime == "" {
		return "", "", cliError("unsupported_generated_file", "Supported output: PDF, PNG, JPEG, WebP, TXT, CSV or JSON with valid content.")
	}
	name = safeFileName.ReplaceAllString(path.Base(name), "_")
	if len(name) > 100 || name == ext || strings.HasPrefix(name, ".") {
		name = "output" + ext
	}
	return name, mime, nil
}
func FileDigest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func (c *AgentHarness) saveGeneratedFiles(ctx context.Context, run entities.AgentRun, jobID string, files []entities.GeneratedFile) ([]entities.StoredFile, error) {
	if c.Files == nil {
		return nil, ErrAgentUnavailable
	}
	if len(files) > 3 {
		return nil, ErrAgentCapacity
	}
	total := 0
	metadata := []entities.StoredFile{}
	// Validate the whole output set before committing any content.
	for i, file := range files {
		name, mime, err := GeneratedFileType(file.Name, file.Data)
		if err != nil {
			return nil, err
		}
		total += len(file.Data)
		if total > MaxGeneratedTotalBytes {
			return nil, ErrAgentCapacity
		}
		metadata = append(metadata, entities.StoredFile{ID: stableID("file", fmt.Sprintf("%s:%d", jobID, i)), Name: name, MIME: mime, Size: len(file.Data), SHA256: FileDigest(file.Data), Chunks: (len(file.Data) + FileChunkBytes - 1) / FileChunkBytes, RunID: run.ID, ThreadID: run.ThreadID, CreatedAt: time.Now().UTC()})
	}
	for i, meta := range metadata {
		if err := c.Files.SaveFile(ctx, meta, files[i].Data); err != nil {
			return metadata[:i], err
		}
	}
	return metadata, nil
}
func (c *AgentHarness) FileMetadata(ctx context.Context, id string) (entities.StoredFile, error) {
	if !ValidFileID(id) {
		return entities.StoredFile{}, ErrInvalidRecord
	}
	if c.Files == nil {
		return entities.StoredFile{}, ErrAgentUnavailable
	}
	return c.Files.FileMetadata(ctx, id)
}
func (c *AgentHarness) FileContent(ctx context.Context, id string) (entities.StoredFile, []byte, error) {
	if !ValidFileID(id) {
		return entities.StoredFile{}, nil, ErrInvalidRecord
	}
	if c.Files == nil {
		return entities.StoredFile{}, nil, ErrAgentUnavailable
	}
	return c.Files.FileContent(ctx, id)
}
