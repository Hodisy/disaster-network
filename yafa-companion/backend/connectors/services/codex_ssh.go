package services

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yafa/backend/entities"
)

// A fixed remote entry point receives structured stdin. Model text is never
// interpolated into the remote shell command, SSH options or filesystem paths.
type CodexSSH struct {
	Host, User, KeyFile, KnownHostsFile, Runner string
	Port                                        int
}

var sshHost = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.:-]{0,252}$`)
var sshUser = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
var sshRunner = regexp.MustCompile(`^/[a-zA-Z0-9_./-]{1,240}$`)

func (c CodexSSH) Configured() bool {
	if !sshHost.MatchString(c.Host) || !sshUser.MatchString(c.User) || c.User == "root" || c.Port < 1 || c.Port > 65535 || !sshRunner.MatchString(c.Runner) || filepath.Clean(c.Runner) != c.Runner {
		return false
	}
	for _, path := range []string{c.KeyFile, c.KnownHostsFile} {
		if !filepath.IsAbs(path) || strings.ContainsAny(path, "\x00\r\n") {
			return false
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return false
		}
		if path == c.KeyFile && info.Mode().Perm()&0077 != 0 {
			return false
		}
	}
	_, err := exec.LookPath("ssh")
	return err == nil
}
func (c CodexSSH) Run(ctx context.Context, in entities.CodexRequest) (entities.CodexResult, error) {
	var zero entities.CodexResult
	if !c.Configured() {
		return zero, errors.New("codex_ssh_not_configured")
	}
	args := []string{"-F", "/dev/null", "-T", "-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "ForwardAgent=no", "-o", "ClearAllForwardings=yes", "-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=8", "-o", "ServerAliveInterval=10", "-o", "ServerAliveCountMax=2", "-o", "PermitLocalCommand=no", "-o", "EscapeChar=none", "-o", "UserKnownHostsFile=" + c.KnownHostsFile, "-i", c.KeyFile, "-p", strconv.Itoa(c.Port), "--", c.User + "@" + c.Host, c.Runner}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=/tmp", "LANG=C.UTF-8"}
	cmd.WaitDelay = 2 * time.Second
	cmd.Stderr = io.Discard
	input, err := cmd.StdinPipe()
	if err != nil {
		return zero, errors.New("codex_ssh_failed")
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		return zero, errors.New("codex_ssh_failed")
	}
	if err = cmd.Start(); err != nil {
		return zero, errors.New("codex_ssh_failed")
	}
	// Keep stdin open while connected. The remote runner treats EOF as a stop.
	defer input.Close()
	if err = json.NewEncoder(input).Encode(in); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return zero, errors.New("codex_ssh_failed")
	}
	scanner := bufio.NewScanner(io.LimitReader(output, 196609))
	scanner.Buffer(make([]byte, 4096), 196608)
	var result *entities.CodexResult
	invalid := false
	for scanner.Scan() {
		var event entities.CodexEvent
		if json.Unmarshal(scanner.Bytes(), &event) != nil || event.Type != "result" || event.Result == nil || result != nil {
			invalid = true
			_ = cmd.Process.Kill()
			break
		}
		result = event.Result
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if invalid || scanErr != nil || waitErr != nil || result == nil || ctx.Err() != nil {
		return zero, errors.New("codex_ssh_outcome_unknown")
	}
	if len(result.Text) > 12000 || len(result.ChangedFiles) > 64 || len(result.ErrorCode) > 80 || result.Usage.InputTokens < 0 || result.Usage.OutputTokens < 0 || result.Usage.CachedInputTokens < 0 || result.Usage.CachedInputTokens > result.Usage.InputTokens || result.Usage.InputTokens > 1000000000 || result.Usage.OutputTokens > 1000000000 {
		return zero, errors.New("codex_invalid_result")
	}
	switch result.Status {
	case "succeeded", "failed", "cancelled", "timed_out":
	default:
		return zero, errors.New("codex_invalid_result")
	}
	for _, path := range result.ChangedFiles {
		if path == "" || len(path) > 240 || filepath.IsAbs(path) || filepath.Clean(path) != path || path == ".." || strings.HasPrefix(path, "../") {
			return zero, errors.New("codex_invalid_result")
		}
	}
	return *result, nil
}
