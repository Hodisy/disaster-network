// Package importbundle reads a private local package; it never contacts a service.
package importbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"yafa/backend/core"
	"yafa/backend/entities"
)

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readFile(root, name string, max int64) ([]byte, error) {
	if filepath.IsAbs(name) || strings.Contains(name, "\\") || filepath.ToSlash(filepath.Clean(name)) != name || name == "." || strings.HasPrefix(name, "../") {
		return nil, errors.New("invalid package path")
	}
	path := root
	for _, component := range strings.Split(name, "/") {
		path = filepath.Join(path, component)
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("package entry unavailable or symlinked")
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("package entry unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > max {
		return nil, errors.New("package entry exceeds bounds")
	}
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil || int64(len(b)) > max {
		return nil, errors.New("package entry read failed")
	}
	return b, nil
}
func Load(root string) (core.PersonalImport, error) {
	var result core.PersonalImport
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return result, errors.New("bundle directory unavailable")
	}
	root = canonical
	manifestBytes, err := readFile(root, "manifest.json", 65536)
	if err != nil {
		return result, err
	}
	var manifest struct {
		Schema   string `json:"schema"`
		ID       string `json:"import_id"`
		AsOf     string `json:"as_of_date"`
		Timezone string `json:"timezone"`
		Files    []struct {
			Path string `json:"path"`
			Size int    `json:"size_bytes"`
			Hash string `json:"sha256"`
		} `json:"files"`
	}
	if json.Unmarshal(manifestBytes, &manifest) != nil || manifest.Schema != "yafa.private-context-bundle.v1" || manifest.Timezone != "Europe/Paris" || len(manifest.Files) == 0 || len(manifest.Files) > 32 {
		return result, errors.New("invalid bundle manifest")
	}
	files := map[string][]byte{}
	total := 0
	for _, entry := range manifest.Files {
		if _, exists := files[entry.Path]; exists {
			return result, errors.New("duplicate package entry")
		}
		body, e := readFile(root, entry.Path, 131072)
		if e != nil {
			return result, e
		}
		if len(body) != entry.Size || digest(body) != entry.Hash {
			return result, errors.New("bundle integrity mismatch")
		}
		total += len(body)
		if total > 1048576 {
			return result, errors.New("bundle too large")
		}
		files[entry.Path] = body
	}
	var pkg struct {
		Schema   string `json:"schema"`
		ID       string `json:"import_id"`
		Timezone string `json:"timezone"`
		Projects []struct {
			Key  string `json:"key"`
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"projects"`
		Conversations []json.RawMessage `json:"conversations"`
		Items         []json.RawMessage `json:"items"`
	}
	if json.Unmarshal(files["tasks.json"], &pkg) != nil || pkg.Schema != "yafa.private-scheduled-import.v1" || pkg.ID != manifest.ID || pkg.Timezone != manifest.Timezone || len(pkg.Conversations) != 0 {
		return result, errors.New("invalid task package")
	}
	result = core.PersonalImport{ID: manifest.ID, Fingerprint: digest(manifestBytes), AsOf: manifest.AsOf, Profile: string(files["context/user.md"]), Directive: string(files["context/directive.md"])}
	for _, p := range pkg.Projects {
		result.Projects = append(result.Projects, core.PersonalProject{Key: p.Key, Name: p.Name, Kind: p.Kind})
	}
	for _, raw := range pkg.Items {
		var item struct {
			Key       string  `json:"local_key"`
			Owner     string  `json:"owner"`
			Kind      string  `json:"content_type"`
			Project   *string `json:"project_ref"`
			Scheduled bool    `json:"scheduled"`
			Status    string  `json:"source_status"`
			Task      struct {
				Title     string  `json:"title"`
				Notes     string  `json:"notes"`
				Date      *string `json:"scheduled_date"`
				Time      *string `json:"scheduled_time"`
				Due       *string `json:"due_at"`
				Completed bool    `json:"completed"`
			} `json:"task"`
			Schedule struct {
				Timezone string `json:"timezone"`
			} `json:"schedule_metadata"`
			Provenance struct {
				CompletedAt *string `json:"completed_at"`
			} `json:"provenance"`
		}
		if json.Unmarshal(raw, &item) != nil || item.Owner != "me" || !item.Scheduled || item.Schedule.Timezone != "Europe/Paris" || (item.Kind != "event" && item.Kind != "task") || (item.Status != "done" && item.Status != "todo" && item.Status != "in_progress") || item.Task.Completed != (item.Status == "done") || item.Provenance.CompletedAt != nil {
			return result, errors.New("invalid scheduled item")
		}
		ref := ""
		if item.Project != nil {
			ref = *item.Project
		}
		result.Items = append(result.Items, core.PersonalItem{Key: item.Key, ProjectRef: ref, Task: core.CreateTaskInput{Title: item.Task.Title, Notes: item.Task.Notes, ScheduledDate: item.Task.Date, ScheduledTime: item.Task.Time, DueAt: item.Task.Due}, Completed: item.Task.Completed, Metadata: raw})
	}
	var catalog struct {
		Schema  string `json:"schema"`
		Entries []struct {
			ID       string   `json:"id"`
			Path     string   `json:"path_in_package"`
			Title    string   `json:"title"`
			Purpose  string   `json:"purpose"`
			AsOf     string   `json:"as_of"`
			Revision string   `json:"package_revision"`
			Source   string   `json:"source_id"`
			Hash     string   `json:"sha256"`
			Size     int      `json:"size_bytes"`
			Sections []string `json:"sections"`
		} `json:"entries"`
	}
	if json.Unmarshal(files["catalog.json"], &catalog) != nil || catalog.Schema != "yafa.private-memory-catalog.v1" || len(catalog.Entries) > 32 {
		return result, errors.New("invalid memory catalog")
	}
	current := false
	for _, doc := range catalog.Entries {
		body, ok := files[doc.Path]
		if !ok || digest(body) != doc.Hash || len(body) != doc.Size {
			return result, errors.New("catalog integrity mismatch")
		}
		if doc.ID == "current-context" {
			if doc.Path != "context/now.md" || doc.AsOf != manifest.AsOf {
				return result, errors.New("invalid current note")
			}
			current = true
		}
		result.Documents = append(result.Documents, entities.PersonalDocument{ID: doc.ID, Title: doc.Title, Purpose: doc.Purpose, AsOf: doc.AsOf, Sections: doc.Sections, Body: string(body), Revision: doc.Revision, SHA256: doc.Hash, SourceID: doc.Source})
	}
	if !current {
		return result, errors.New("current note missing")
	}
	return result, nil
}
