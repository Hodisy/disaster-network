package entities

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Preserve additive fields owned by another compatible runtime slice. This is
// essential for maintenance imports; unknown state must not disappear on save.
func workspaceFields() map[string]bool {
	fields := map[string]bool{}
	t := reflect.TypeOf(Workspace{})
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			fields[name] = true
		}
	}
	return fields
}
func (w *Workspace) UnmarshalJSON(data []byte) error {
	type plain Workspace
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for name := range workspaceFields() {
		delete(fields, name)
	}
	*w = Workspace(value)
	w.ExtraFields = fields
	return nil
}
func (w Workspace) MarshalJSON() ([]byte, error) {
	type plain Workspace
	raw, err := json.Marshal(plain(w))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	known := workspaceFields()
	for name, value := range w.ExtraFields {
		if !known[name] {
			fields[name] = value
		}
	}
	return json.Marshal(fields)
}
