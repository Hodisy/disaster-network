// Thin terminal client for the real Go API. No model key or duplicate agent loop.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"yafa/backend/entities"
)

func main() {
	token := os.Getenv("YAFA_API_TOKEN")
	base := strings.TrimRight(os.Getenv("YAFA_API_URL"), "/")
	if base == "" {
		base = "http://127.0.0.1:8088"
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "Set the private YAFA_API_TOKEN environment variable. No OpenAI key belongs in this client.")
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	client := &http.Client{Timeout: 10 * time.Second}
	request := func(callCtx context.Context, method, path string, value, out any) error {
		var body []byte
		if value != nil {
			body, _ = json.Marshal(value)
		}
		req, err := http.NewRequestWithContext(callCtx, method, base+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("API connection failed; no request was replayed")
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			var failure struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			_ = json.NewDecoder(io.LimitReader(response.Body, 8192)).Decode(&failure)
			return fmt.Errorf("HTTP %d: %s %s", response.StatusCode, failure.Error, failure.Message)
		}
		return json.NewDecoder(io.LimitReader(response.Body, 524288)).Decode(out)
	}
	fmt.Println("Yafa backend chat. /usage, /status, /quit. Requests use the real backend and may incur API costs when enabled.")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 8192)
	thread := ""
	lines := make(chan string)
	go func() {
		defer close(lines)
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		fmt.Print("You > ")
		var line string
		select {
		case value, ok := <-lines:
			if !ok {
				return
			}
			line = value
		case <-ctx.Done():
			return
		}
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if text == "/quit" {
			return
		}
		if text == "/usage" || text == "/status" {
			path := "/v1/safety"
			if text == "/status" {
				path = "/v1/agent/status"
			}
			var status any
			if err := request(ctx, http.MethodGet, path, nil, &status); err != nil {
				fmt.Println(err)
			} else {
				raw, _ := json.MarshalIndent(status, "", "  ")
				fmt.Println(string(raw))
			}
			continue
		}
		if thread == "" {
			var created entities.Thread
			if err := request(ctx, http.MethodPost, "/v1/threads", map[string]string{"event_id": fmt.Sprintf("terminal-%d", time.Now().UnixNano()), "origin": "desktop"}, &created); err != nil {
				fmt.Println(err)
				continue
			}
			thread = created.ID
		}
		var run entities.AgentRun
		if err := request(ctx, http.MethodPost, "/v1/agent/runs", map[string]string{"thread_id": thread, "event_id": fmt.Sprintf("message-%d", time.Now().UnixNano()), "message": text}, &run); err != nil {
			fmt.Println(err)
			continue
		}
		for run.Status == "queued" || run.Status == "running" {
			select {
			case <-ctx.Done():
				cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				var result any
				_ = request(cleanup, http.MethodPost, "/v1/agent/runs/"+run.ID+"/cancel", nil, &result)
				cancel()
				return
			case <-time.After(500 * time.Millisecond):
			}
			if err := request(ctx, http.MethodGet, "/v1/agent/runs/"+run.ID, nil, &run); err != nil {
				fmt.Println(err)
				break
			}
		}
		fmt.Println("Yafa >", run.Output)
		fmt.Printf("[%s; %d input / %d output tokens; known estimate $%.6f; cost incomplete=%t; error=%s]\n", run.Status, run.Usage.InputTokens, run.Usage.OutputTokens, run.Usage.KnownCostUSD, run.Usage.Incomplete || run.Usage.AdditionalCostsUnpriced, run.ErrorCode)
	}
}
