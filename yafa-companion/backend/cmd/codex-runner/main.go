// Fixed remote SSH entry point. Install under a dedicated unprivileged account,
// authenticate Codex there, and keep application/SSH/admin secrets out of it.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
	"yafa/backend/entities"
)

func main() {
	result := run()
	_ = json.NewEncoder(os.Stdout).Encode(entities.CodexEvent{Type: "result", Result: &result})
}
func failure(code string) entities.CodexResult {
	return entities.CodexResult{Status: "failed", ErrorCode: code, ChangedFiles: []string{}}
}
func bounded(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n]
}
func run() entities.CodexResult {
	if os.Geteuid() == 0 {
		return failure("codex_runner_requires_unprivileged_user")
	}
	// One bounded JSON line, then a live stdin lease: EOF cancels the child.
	input := bufio.NewReaderSize(os.Stdin, 32768)
	line, err := input.ReadSlice('\n')
	if err != nil {
		return failure("codex_invalid_request")
	}
	var in entities.CodexRequest
	d := json.NewDecoder(strings.NewReader(string(line)))
	d.DisallowUnknownFields()
	if d.Decode(&in) != nil || d.Decode(new(any)) != io.EOF || !regexp.MustCompile(`^codex-[a-f0-9]{32}$`).MatchString(in.JobID) || len(in.Task) < 1 || len(in.Task) > 3000 || !utf8.ValidString(in.Task) || strings.ContainsRune(in.Task, 0) || in.MaxSeconds < 30 || in.MaxSeconds > 600 {
		return failure("codex_invalid_request")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return failure("codex_runner_configuration")
	}
	root := os.Getenv("YAFA_CODEX_WORK_ROOT")
	if root == "" {
		root = filepath.Join(home, "jobs")
	}
	auth := os.Getenv("CODEX_HOME")
	if auth == "" {
		auth = filepath.Join(home, ".codex")
	}
	binary := os.Getenv("YAFA_CODEX_BINARY")
	if binary == "" {
		binary = "/usr/local/bin/codex"
	}
	if !filepath.IsAbs(root) || !filepath.IsAbs(auth) || !filepath.IsAbs(binary) || root == "/" || strings.HasPrefix(auth, root+"/") || strings.HasPrefix(root, auth+"/") || auth == root {
		return failure("codex_runner_configuration")
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return failure("codex_workspace_unavailable")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return failure("codex_workspace_unavailable")
	}
	auth, err = filepath.EvalSymlinks(auth)
	if err != nil {
		return failure("codex_authentication_unavailable")
	}
	dir := filepath.Join(root, in.JobID)
	// Exclusive creation prevents a repeated SSH request from replaying a job.
	if err = os.Mkdir(dir, 0700); err != nil {
		return failure("codex_workspace_exists_or_unavailable")
	}
	tmp := filepath.Join(dir, ".tmp")
	if os.Mkdir(tmp, 0700) != nil {
		return failure("codex_workspace_unavailable")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(in.MaxSeconds)*time.Second)
	defer cancel()
	go func() { _, _ = io.Copy(io.Discard, input); cancel() }()
	// Official workspace-only permission profile: deny outside reads, keep only
	// minimal system tools and this job writable. Credentials stay CLI-internal.
	filesystem := `{":root"="deny",":minimal"="read",":tmpdir"="write",":slash_tmp"="deny",` + strconv.Quote(dir) + `="write",` + strconv.Quote(auth) + `="deny"}`
	args := []string{"--ask-for-approval", "never", "-c", `default_permissions="yafa-job"`, "-c", `permissions.yafa-job.extends=":workspace"`, "-c", "permissions.yafa-job.filesystem=" + filesystem, "exec", "--json", "--color", "never", "--ignore-user-config", "--ephemeral", "--skip-git-repo-check", "-C", dir, "-"}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Env = []string{"PATH=/usr/local/bin:/usr/bin:/bin", "HOME=" + dir, "CODEX_HOME=" + auth, "TMPDIR=" + tmp, "LANG=C.UTF-8"}
	cmd.Stdin = strings.NewReader("Work only in the assigned job workspace. Do not deploy, access other projects, read credentials, contact Yafa's administrative API, or create callbacks. Return a concise result and relative paths of created files.\n\nUser task:\n" + in.Task)
	cmd.Stderr = io.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return failure("codex_runner_failed")
	}
	if err = cmd.Start(); err != nil {
		return failure("codex_runner_failed")
	}
	// Do not store raw JSONL (reasoning, command output or arbitrary metadata).
	result := entities.CodexResult{Status: "failed", ChangedFiles: []string{}}
	scanner := bufio.NewScanner(io.LimitReader(stdout, 4*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 262144)
	total := 0
	invalid := false
	completed := false
	for scanner.Scan() {
		total += len(scanner.Bytes()) + 1
		if total > 4*1024*1024 {
			invalid = true
			cancel()
			break
		}
		var event struct {
			Type  string `json:"type"`
			Usage struct {
				Input  int64 `json:"input_tokens"`
				Cached int64 `json:"cached_input_tokens"`
				Output int64 `json:"output_tokens"`
			} `json:"usage"`
			Item struct {
				Type    string `json:"type"`
				Text    string `json:"text"`
				Changes []struct {
					Path string `json:"path"`
				} `json:"changes"`
			} `json:"item"`
		}
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			invalid = true
			cancel()
			break
		}
		switch event.Type {
		case "turn.completed":
			completed = true
			result.Usage = entities.CodexUsage{InputTokens: event.Usage.Input, CachedInputTokens: event.Usage.Cached, OutputTokens: event.Usage.Output, Confirmed: true}
		case "turn.failed", "error":
			result.ErrorCode = "codex_turn_failed"
		case "item.completed":
			if event.Item.Type == "agent_message" {
				result.Text = bounded(event.Item.Text, 12000)
			}
			if event.Item.Type == "file_change" {
				for _, change := range event.Item.Changes {
					p := change.Path
					if filepath.IsAbs(p) {
						p, _ = filepath.Rel(dir, p)
					}
					p = filepath.Clean(p)
					if p != "." && p != ".." && !strings.HasPrefix(p, "../") && len(p) <= 240 && len(result.ChangedFiles) < 64 {
						result.ChangedFiles = append(result.ChangedFiles, p)
					}
				}
			}
		}
	}
	if scanner.Err() != nil {
		invalid = true
		cancel()
	}
	err = cmd.Wait()
	// Kill any remaining descendants in this process group, even on normal exit.
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Status = "timed_out"
			result.ErrorCode = "codex_deadline"
		} else {
			result.Status = "cancelled"
			result.ErrorCode = "codex_cancelled"
		}
	} else if invalid {
		result.ErrorCode = "codex_output_limit_or_invalid"
	} else if err != nil {
		result.ErrorCode = "codex_process_failed"
	} else if completed && result.ErrorCode == "" {
		result.Status = "succeeded"
	} else {
		result.ErrorCode = "codex_result_incomplete"
	}
	return result
}
