package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yafa/backend/entities"
)

type TokenPrices struct{ Input, CachedInput, Output float64 }
type OpenAIResponses struct {
	APIKey string
	Model  string
	Client *http.Client
	Prices *TokenPrices
}

func (c OpenAIResponses) Configured() bool { return c.APIKey != "" && c.Model != "" && c.Client != nil }
func (c OpenAIResponses) Name() string     { return c.Model }
func (c OpenAIResponses) Respond(ctx context.Context, input entities.ModelRequest) (entities.ModelReply, error) {
	if !c.Configured() {
		return entities.ModelReply{}, &entities.ModelFailure{Code: "model_not_configured"}
	}
	body, err := json.Marshal(map[string]any{"model": c.Model, "instructions": input.Instructions, "input": input.Input, "tools": input.Tools, "max_output_tokens": input.MaxOutputTokens, "parallel_tool_calls": false, "store": false, "include": []string{"reasoning.encrypted_content"}})
	if err != nil {
		return entities.ModelReply{}, &entities.ModelFailure{Code: "invalid_model_request"}
	}
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(body))
		if err != nil {
			return entities.ModelReply{}, &entities.ModelFailure{Code: "invalid_model_request"}
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		response, err := c.Client.Do(req)
		if err != nil {
			return entities.ModelReply{}, &entities.ModelFailure{Code: "model_connection_failed", UsageUnknown: true}
		}
		if response.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 8192))
			response.Body.Close()
			delay := 1
			if n, e := strconv.Atoi(response.Header.Get("Retry-After")); e == nil {
				delay = max(1, min(n, 2))
			}
			select {
			case <-ctx.Done():
				return entities.ModelReply{}, ctx.Err()
			case <-time.After(time.Duration(delay) * time.Second):
			}
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(response.Body, 524289))
		response.Body.Close()
		if readErr != nil || len(data) > 524288 {
			return entities.ModelReply{}, &entities.ModelFailure{Code: "invalid_model_response", UsageUnknown: true}
		}
		if response.StatusCode != http.StatusOK {
			code := "model_request_rejected"
			switch response.StatusCode {
			case 400:
				code = "invalid_model_request"
			case 401:
				code = "openai_authentication_failed"
			case 403:
				code = "openai_access_denied"
			case 404:
				code = "model_unavailable"
			case 429:
				code = "model_rate_limited"
			default:
				if response.StatusCode >= 500 {
					code = "model_provider_unavailable"
				}
			}
			// Never expose provider response bodies: they can echo private request data.
			return entities.ModelReply{}, &entities.ModelFailure{Code: code, UsageUnknown: response.StatusCode >= 500}
		}
		var raw struct {
			ID     string            `json:"id"`
			Status string            `json:"status"`
			Output []json.RawMessage `json:"output"`
			Usage  *struct {
				Input        int64 `json:"input_tokens"`
				Output       int64 `json:"output_tokens"`
				InputDetails struct {
					Cached int64 `json:"cached_tokens"`
				} `json:"input_tokens_details"`
				OutputDetails struct {
					Reasoning int64 `json:"reasoning_tokens"`
				} `json:"output_tokens_details"`
			} `json:"usage"`
		}
		if json.Unmarshal(data, &raw) != nil {
			return entities.ModelReply{}, &entities.ModelFailure{Code: "invalid_model_response", UsageUnknown: true}
		}
		reply := entities.ModelReply{ID: raw.ID, Status: raw.Status, Output: raw.Output}
		if raw.Usage == nil {
			reply.Usage.Incomplete = true
		} else {
			u := raw.Usage
			reply.Usage = entities.Usage{InputTokens: u.Input, OutputTokens: u.Output, CachedTokens: u.InputDetails.Cached, ReasoningTokens: u.OutputDetails.Reasoning}
			if u.Input < 0 || u.Output < 0 || u.Input > 1000000000 || u.Output > 1000000000 || u.InputDetails.Cached < 0 || u.InputDetails.Cached > u.Input || u.OutputDetails.Reasoning < 0 || u.OutputDetails.Reasoning > u.Output {
				reply.Usage = entities.Usage{Incomplete: true}
			}
			if c.Prices != nil && !reply.Usage.Incomplete {
				cost := (float64(u.Input-u.InputDetails.Cached)*c.Prices.Input + float64(u.InputDetails.Cached)*c.Prices.CachedInput + float64(u.Output)*c.Prices.Output) / 1000000
				reply.Usage.EstimatedCostUSD = &cost
			}
		}
		for _, data := range raw.Output {
			var item struct {
				Type      string `json:"type"`
				CallID    string `json:"call_id"`
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
				Content   []struct {
					Type    string `json:"type"`
					Text    string `json:"text"`
					Refusal string `json:"refusal"`
				} `json:"content"`
			}
			if json.Unmarshal(data, &item) != nil {
				reply.Status = "failed"
				continue
			}
			if item.Type == "function_call" {
				if item.CallID == "" || len(item.CallID) > 128 || len(item.Name) > 80 || len(item.Arguments) > 8192 || len(reply.Calls) >= 32 {
					reply.Status = "failed"
					continue
				}
				reply.Calls = append(reply.Calls, entities.ToolCall{CallID: item.CallID, Name: item.Name, Arguments: item.Arguments})
			}
			if item.Type == "message" {
				for _, content := range item.Content {
					if content.Type == "output_text" {
						reply.Text += content.Text
					}
					if content.Type == "refusal" {
						reply.Text += content.Refusal
					}
				}
			}
		}
		reply.Text = strings.TrimSpace(reply.Text)
		return reply, nil
	}
	return entities.ModelReply{}, &entities.ModelFailure{Code: "model_rate_limited"}
}
