package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"yafa/backend/core"
	"yafa/backend/entities"
)

type MemoryHTTP struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func (r MemoryHTTP) request(ctx context.Context, method, path string, body []byte, output any) error {
	if r.Client == nil {
		return core.ErrMemoryUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(r.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return core.ErrMemoryUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+r.Token)
	req.Header.Set("Content-Type", "application/json")
	response, err := r.Client.Do(req)
	if err != nil {
		return core.ErrMemoryUnavailable
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusNotFound:
		return core.ErrRecordNotFound
	case http.StatusConflict:
		return core.ErrVersionConflict
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return core.ErrInvalidRecord
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: upstream status %d", core.ErrMemoryUnavailable, response.StatusCode)
	}
	if output != nil {
		if json.NewDecoder(io.LimitReader(response.Body, 1048576)).Decode(output) != nil {
			return core.ErrMemoryUnavailable
		}
	}
	return nil
}
func (r MemoryHTTP) Ready(ctx context.Context) error {
	return r.request(ctx, http.MethodGet, "/health/ready", nil, nil)
}
func (r MemoryHTTP) Read(ctx context.Context, key string) (entities.MemoryRecord, error) {
	var result entities.MemoryRecord
	err := r.request(ctx, http.MethodGet, "/v1/records/"+url.PathEscape(key), nil, &result)
	return result, err
}
func (r MemoryHTTP) Write(ctx context.Context, key, expected string, value json.RawMessage) (entities.MemoryRecord, error) {
	body, err := json.Marshal(struct {
		ExpectedVersion string          `json:"expected_version"`
		Value           json.RawMessage `json:"value"`
	}{expected, value})
	if err != nil {
		return entities.MemoryRecord{}, core.ErrInvalidRecord
	}
	var result entities.MemoryRecord
	err = r.request(ctx, http.MethodPost, "/v1/records/"+url.PathEscape(key), body, &result)
	return result, err
}
