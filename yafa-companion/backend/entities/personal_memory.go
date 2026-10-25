package entities

import (
	"encoding/json"
	"time"
)

// Personal documents are owner data, never executable paths or tool grants.
type PersonalDocument struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Purpose  string   `json:"purpose"`
	AsOf     string   `json:"as_of"`
	Sections []string `json:"sections"`
	Body     string   `json:"body"`
	Revision string   `json:"revision"`
	SHA256   string   `json:"sha256"`
	SourceID string   `json:"source_id"`
}
type PersonalImportReceipt struct {
	Fingerprint string                     `json:"fingerprint"`
	ImportedAt  time.Time                  `json:"imported_at"`
	ProjectIDs  map[string]string          `json:"project_ids"`
	TaskIDs     map[string]string          `json:"task_ids"`
	Metadata    map[string]json.RawMessage `json:"metadata"`
}
