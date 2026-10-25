package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"strings"
	"sync"
	"time"
	"yafa/backend/entities"
	"yafa/backend/services"
)

type VoiceStart struct {
	Type          string `json:"type"`
	ThreadID      string `json:"thread_id"`
	Surface       string `json:"surface"`
	UserInitiated bool   `json:"user_initiated"`
}
type VoiceSessions struct {
	Agent              *AgentHarness
	Provider           services.LiveProvider
	Model              string
	Voice              string
	Root               context.Context
	MaxDurationSeconds int
	gate               chan struct{}
	mu                 sync.Mutex
	activeID           string
	activeCancel       context.CancelFunc
	activeStopReason   string
	IdleSeconds        int
	USDPerMinute       *float64
}

func NewVoiceSessions(agent *AgentHarness, provider services.LiveProvider, model, voice string, root context.Context, maxSeconds int) *VoiceSessions {
	return &VoiceSessions{Agent: agent, Provider: provider, Model: model, Voice: voice, Root: root, MaxDurationSeconds: maxSeconds, gate: make(chan struct{}, 1)}
}
func (c *VoiceSessions) Configured() bool {
	return c != nil && c.Agent != nil && c.Agent.Features.Live && c.Agent.Configured() && c.Provider != nil && c.Provider.Configured() && c.Model != "" && c.Voice != ""
}

type voicePacket struct {
	data json.RawMessage
	err  error
}
type delegatedResult struct {
	generation int
	id         string
	run        entities.AgentRun
	err        error
}

func (c *VoiceSessions) Run(requestCtx context.Context, peer services.VoicePeer, start VoiceStart) (returnErr error) {
	if c.Root.Err() != nil {
		return c.Root.Err()
	}
	if !c.Configured() {
		return ErrAgentUnavailable
	}
	if start.Type != "session.start" || !start.UserInitiated || !validAgentID(start.ThreadID, "thread") || (start.Surface != "mobile" && start.Surface != "desktop") {
		return ErrInvalidRecord
	}
	thread, err := c.Agent.Thread(requestCtx, start.ThreadID)
	if err != nil {
		return err
	}
	select {
	case c.gate <- struct{}{}:
		defer func() { <-c.gate }()
	default:
		return ErrAgentBusy
	}
	ctx, cancel := context.WithCancel(requestCtx)
	defer cancel()
	stopRoot := context.AfterFunc(c.Root, cancel)
	defer stopRoot()
	// Keep provider reads alive during the bounded graceful-close period.
	transportCtx, stopTransport := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(c.MaxDurationSeconds+30)*time.Second)
	defer stopTransport()
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	record := entities.VoiceRecord{ID: "voice-" + hex.EncodeToString(bytes), ThreadID: start.ThreadID, Surface: start.Surface, Model: c.Model, Status: "starting", StartedAt: time.Now().UTC(), Transcripts: []entities.VoiceTranscript{}, RunIDs: []string{}}
	prompt, err := c.Agent.BuildPrompt(ctx, thread, "live", record.StartedAt, c.Model)
	if err != nil {
		return err
	}
	if c.Agent.Safety != nil {
		record.BudgetDate, err = c.Agent.Safety.Reserve(ctx, record.ID, "voice", 0, float64(c.MaxDurationSeconds+30))
		if err != nil {
			return err
		}
	}
	record.PromptHash, record.PromptDocuments = prompt.Hash, prompt.Documents
	if err = c.Agent.savePrompt(ctx, record.ID, prompt); err != nil {
		return err
	}
	version, err := c.Agent.Repository.SaveVoice(ctx, "0", record)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.activeID = record.ID
	c.activeCancel = cancel
	c.activeStopReason = ""
	c.mu.Unlock()
	activeRun := ""
	emit := func(value any) {
		sendCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer done()
		if peer.Send(sendCtx, value) != nil {
			cancel()
		}
	}
	defer func() {
		if activeRun != "" {
			c.Agent.Cancel(activeRun)
		}
		cleanup, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer done()
		if _, latest, e := c.Agent.Repository.LoadVoice(cleanup, record.ID); e == nil {
			version = latest
		}
		now := time.Now().UTC()
		record.FinishedAt = &now
		record.WallSeconds = now.Sub(record.StartedAt).Seconds()
		if record.UsageFinalConfirmed && c.USDPerMinute != nil {
			cost := record.DurationSeconds / 60 * (*c.USDPerMinute)
			record.EstimatedCostUSD = &cost
		}
		if returnErr != nil {
			record.ErrorCode = agentErrorCode(returnErr)
			record.Status = "failed"
		} else {
			record.Status = "closed"
		}
		if c.Agent.Safety != nil && record.BudgetDate != "" {
			known := 0.0
			if record.EstimatedCostUSD != nil {
				known = *record.EstimatedCostUSD
			}
			record.BudgetSettlementConfirmed = c.Agent.Safety.Settle(cleanup, record.BudgetDate, record.ID, record.Status, 0, record.DurationSeconds, known, record.UsageFinalConfirmed, record.EstimatedCostUSD != nil, record.ErrorCode) == nil
		}

		_, saveErr := c.Agent.Repository.SaveVoice(cleanup, version, record)
		emit(map[string]any{"type": "voice.closed", "session_id": record.ID, "duration_seconds": record.DurationSeconds, "finalized": record.Finalized, "usage_final_confirmed": record.UsageFinalConfirmed, "persistence_confirmed": saveErr == nil, "error": record.ErrorCode})
		_ = peer.Close()
		c.mu.Lock()
		c.activeID = ""
		c.activeCancel = nil
		c.mu.Unlock()
	}()
	if c.Agent.Safety != nil {
		if err := c.Agent.Safety.Check(ctx); err != nil {
			return err
		}
	}
	provider, err := c.Provider.Open(ctx)
	if err != nil {
		return err
	}
	defer provider.Close()
	providerEvents := make(chan voicePacket, 8)
	clientEvents := make(chan voicePacket, 8)
	go func() {
		for {
			data, err := provider.Receive(transportCtx)
			select {
			case providerEvents <- voicePacket{data, err}:
			case <-transportCtx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	go func() {
		for {
			data, err := peer.Receive(ctx)
			select {
			case clientEvents <- voicePacket{data, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	config := entities.VoiceConfig{Model: c.Model, Instructions: prompt.Instructions, Audio: map[string]any{"format": map[string]any{"type": "audio/pcm", "rate": 24000}, "output": map[string]any{"voice": c.Voice}}, Delegation: map[string]any{"type": "client"}}
	if err = provider.Send(ctx, encoded(map[string]any{"type": "session.start", "session": config})); err != nil {
		return &entities.ModelFailure{Code: "live_start_failed", UsageUnknown: true}
	}
	started := false
	closing := false
	muted := false
	generation := 0
	dirty := false
	delegations := map[string]bool{}
	pending := ""
	results := make(chan delegatedResult, 8)
	quiet := time.NewTimer(time.Hour)
	quiet.Stop()
	defer quiet.Stop()
	startTimer := time.NewTimer(15 * time.Second)
	defer startTimer.Stop()
	idleSeconds := c.IdleSeconds
	if idleSeconds <= 0 {
		idleSeconds = 60
	}
	idleTimer := time.NewTimer(time.Duration(idleSeconds) * time.Second)
	defer idleTimer.Stop()
	warningTimer := time.NewTimer(time.Duration(max(1, c.MaxDurationSeconds-30)) * time.Second)
	defer warningTimer.Stop()
	maxTimer := time.NewTimer(time.Duration(c.MaxDurationSeconds) * time.Second)
	defer maxTimer.Stop()
	closeTimer := time.NewTimer(time.Hour)
	closeTimer.Stop()
	defer closeTimer.Stop()
	persistTimer := time.NewTicker(time.Second)
	defer persistTimer.Stop()
	doneChannel := ctx.Done()
	audioCredit := 24000.0
	lastAudio := time.Now()
	closeSession := func(code string) {
		if closing {
			return
		}
		closing = true
		pending = ""
		quiet.Stop()
		record.ErrorCode = code
		record.CloseReason = code
		if code == "" {
			record.CloseReason = "close_requested"
		}
		emit(map[string]any{"type": "voice.closing", "reason": record.CloseReason})
		if activeRun != "" {
			c.Agent.Cancel(activeRun)
			activeRun = ""
		}
		generation++
		if started {
			sendCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			_ = provider.Send(sendCtx, encoded(map[string]any{"type": "session.close"}))
			done()
		}
		closeTimer.Reset(15 * time.Second)
	}
	for {
		select {
		case <-doneChannel:
			doneChannel = nil
			reason := "client_disconnected"
			c.mu.Lock()
			if c.activeStopReason != "" {
				reason = c.activeStopReason
			}
			c.mu.Unlock()
			if c.Root.Err() != nil {
				reason = "server_shutdown"
			}
			closeSession(reason)
		case <-warningTimer.C:
			if !closing {
				emit(map[string]any{"type": "voice.limit_warning", "remaining_seconds": 30})
			}
		case <-idleTimer.C:
			if activeRun != "" || pending != "" {
				idleTimer.Reset(time.Duration(idleSeconds) * time.Second)
			} else {
				closeSession("voice_idle_limit")
			}
		case <-maxTimer.C:
			closeSession("voice_duration_limit")
		case <-startTimer.C:
			if !started {
				return &entities.ModelFailure{Code: "live_start_timeout", UsageUnknown: true}
			}
		case <-closeTimer.C:
			return &entities.ModelFailure{Code: "live_finalization_unconfirmed", UsageUnknown: true}
		case <-persistTimer.C:
			if dirty {
				saveCtx, finish := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
				next, e := c.Agent.Repository.SaveVoice(saveCtx, version, record)
				finish()
				if e != nil {
					closeSession("voice_storage_unavailable")
				} else {
					version = next
					dirty = false
				}
			}
		case packet := <-clientEvents:
			if packet.err != nil {
				clientEvents = nil
				closeSession("client_disconnected")
				continue
			}
			if closing {
				continue
			}
			var input struct {
				Type  string `json:"type"`
				Audio string `json:"audio,omitempty"`
			}
			if decodeVoiceInput(packet.data, &input) != nil {
				closeSession("invalid_client_event")
				continue
			}
			switch input.Type {
			case "audio.append":
				if !started || muted {
					continue
				}
				data, e := base64.StdEncoding.DecodeString(input.Audio)
				if e != nil || len(data) == 0 || len(data) > 9600 || len(data)%2 != 0 {
					closeSession("invalid_audio_chunk")
					continue
				}
				now := time.Now()
				audioCredit = math.Min(24000, audioCredit+now.Sub(lastAudio).Seconds()*48000)
				lastAudio = now
				if float64(len(data)) > audioCredit {
					closeSession("audio_rate_limit")
					continue
				}
				audioCredit -= float64(len(data))
				if e = provider.Send(ctx, encoded(map[string]any{"type": "session.input_audio.append", "audio": input.Audio})); e != nil {
					return &entities.ModelFailure{Code: "live_transport_failed", UsageUnknown: true}
				}
			case "audio.mute", "audio.unmute":
				muted = input.Type == "audio.mute"
				kind := "session.input_audio.unmute"
				if muted {
					kind = "session.input_audio.mute"
				}
				if started {
					if e := provider.Send(ctx, encoded(map[string]any{"type": kind})); e != nil {
						closeSession("live_transport_failed")
					}
				}
				emit(map[string]any{"type": "voice.relay_mute", "muted": muted})
			case "session.close":
				closeSession("")
			case "run.cancel":
				if activeRun != "" {
					c.Agent.Cancel(activeRun)
				}
				generation++
				activeRun = ""
				pending = ""
				quiet.Stop()
				emit(map[string]any{"type": "run.cancellation_requested"})
			default:
				closeSession("client_event_not_allowed")
			}
		case packet := <-providerEvents:
			if packet.err != nil {
				return &entities.ModelFailure{Code: "live_connection_lost", UsageUnknown: !record.UsageFinalConfirmed}
			}
			var event struct {
				Type    string  `json:"type"`
				Reason  string  `json:"reason"`
				Delta   string  `json:"delta"`
				StartMS float64 `json:"start_ms"`
				EndMS   float64 `json:"end_ms"`
				Usage   *struct {
					Seconds *float64 `json:"seconds"`
				} `json:"usage"`
				ContextWindow struct {
					Ratio float64 `json:"usage_ratio"`
				} `json:"context_window"`
				Delegation struct {
					ID     string `json:"id"`
					Target string `json:"target"`
				} `json:"delegation"`
			}
			if json.Unmarshal(packet.data, &event) != nil {
				return &entities.ModelFailure{Code: "invalid_live_event", UsageUnknown: true}
			}
			switch event.Type {
			case "session.started":
				started = true
				startTimer.Stop()
				record.Status = "active"
				dirty = true
				emit(map[string]any{"type": "voice.ready", "session_id": record.ID, "thread_id": record.ThreadID, "format": "pcm16", "sample_rate": 24000, "channels": 1})
				if closing {
					sendCtx, finish := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
					_ = provider.Send(sendCtx, encoded(map[string]any{"type": "session.close"}))
					finish()
				}
			case "session.output_audio.delta":
				if !closing {
					idleTimer.Reset(time.Duration(idleSeconds) * time.Second)
					emit(map[string]any{"type": "audio.delta", "audio": event.Delta})
				}
			case "session.input_transcript.delta", "session.output_transcript.delta":
				if closing {
					continue
				}
				idleTimer.Reset(time.Duration(idleSeconds) * time.Second)
				role := "assistant"
				if event.Type == "session.input_transcript.delta" {
					role = "user"
				}
				if !utf8Text(event.Delta) || len(event.Delta) > 8192 {
					closeSession("invalid_transcript")
					continue
				}
				if role == "user" {
					generation++
					if activeRun != "" {
						c.Agent.Cancel(activeRun)
						activeRun = ""
					}
					if pending != "" {
						quiet.Reset(600 * time.Millisecond)
					}
				}
				n := len(record.Transcripts)
				if n > 0 && record.Transcripts[n-1].Role == role {
					record.Transcripts[n-1].Text += event.Delta
					record.Transcripts[n-1].EndMS = event.EndMS
				} else {
					record.Transcripts = append(record.Transcripts, entities.VoiceTranscript{Role: role, Text: event.Delta, StartMS: event.StartMS, EndMS: event.EndMS})
				}
				if len(record.Transcripts) > 256 || len(encoded(record.Transcripts)) > 49152 {
					closeSession("voice_context_limit")
					continue
				}
				dirty = true
				emit(map[string]any{"type": "transcript.delta", "role": role, "delta": event.Delta})
			case "session.delegation.created":
				if closing || event.Delegation.Target != "client" || event.Delegation.ID == "" || len(event.Delegation.ID) > 128 || delegations[event.Delegation.ID] {
					continue
				}
				if len(delegations) >= 12 {
					closeSession("voice_delegation_limit")
					continue
				}
				delegations[event.Delegation.ID] = true
				if activeRun != "" {
					c.Agent.Cancel(activeRun)
					activeRun = ""
				}
				generation++
				pending = event.Delegation.ID
				quiet.Reset(600 * time.Millisecond)
			case "session.usage.updated", "session.closed":
				validUsage := event.Usage != nil && event.Usage.Seconds != nil && !math.IsNaN(*event.Usage.Seconds) && !math.IsInf(*event.Usage.Seconds, 0) && *event.Usage.Seconds >= record.DurationSeconds && *event.Usage.Seconds <= 86400
				if validUsage {
					record.DurationSeconds = *event.Usage.Seconds
					dirty = true
				}
				if event.Type == "session.closed" {
					record.Finalized = true
					if record.CloseReason == "" {
						record.CloseReason = event.Reason
					}
					record.UsageFinalConfirmed = validUsage
					return nil
				}
				if event.ContextWindow.Ratio > 0.85 {
					closeSession("voice_context_limit")
				}
				emit(map[string]any{"type": "voice.usage", "seconds": record.DurationSeconds, "final": false})
			case "error":
				closeSession("live_provider_error")
			}
		case <-quiet.C:
			if pending == "" || closing {
				continue
			}
			delegationID := pending
			pending = ""
			text := voiceContext(record.Transcripts)
			if text == "" {
				_ = provider.Send(ctx, encoded(map[string]any{"type": "session.commentary.append", "event_id": stableID("result", record.ID+delegationID), "delegation_id": delegationID, "content": "No usable request transcript is available yet. Ask the user to repeat the request; no action was taken."}))
				continue
			}
			run, e := c.Agent.Submit(ctx, SubmitRun{ThreadID: record.ThreadID, EventID: stableID("voice", record.ID+delegationID), Message: text})
			if e != nil {
				_ = provider.Send(ctx, encoded(map[string]any{"type": "session.commentary.append", "event_id": stableID("result", record.ID+delegationID), "delegation_id": delegationID, "content": "The request was not started. Backend status: " + agentErrorCode(e) + ". Do not claim an action succeeded."}))
				continue
			}
			activeRun = run.ID
			record.RunIDs = append(record.RunIDs, run.ID)
			dirty = true
			currentGeneration := generation
			emit(map[string]any{"type": "run.started", "run_id": run.ID, "thread_id": record.ThreadID})
			go func() {
				value, e := c.Agent.Wait(ctx, run.ID)
				select {
				case results <- delegatedResult{currentGeneration, delegationID, value, e}:
				case <-ctx.Done():
				}
			}()
		case result := <-results:
			if result.generation != generation || closing {
				continue
			}
			activeRun = ""
			content := "Backend result unavailable. No success has been confirmed."
			if result.err == nil {
				content = "Backend run status: " + result.run.Status + ". " + result.run.Output
				if result.run.Status != "completed" {
					content = "Backend run status: " + result.run.Status + "; " + result.run.ErrorCode + ". Some actions may already have committed; read the current state before retrying."
				}
			}
			if e := provider.Send(ctx, encoded(map[string]any{"type": "session.commentary.append", "event_id": stableID("result", record.ID+result.id), "delegation_id": result.id, "content": clipText(content, 400)})); e != nil {
				closeSession("live_transport_failed")
			}
			emit(map[string]any{"type": "run.finished", "run_id": result.run.ID, "status": result.run.Status})
		}
	}
}
func (c *VoiceSessions) Shutdown(ctx context.Context) {
	for len(c.gate) != 0 {
		select {
		case <-ctx.Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
}
func utf8Text(value string) bool {
	return json.Valid(encoded(value)) && strings.ToValidUTF8(value, "") == value
}
func voiceContext(parts []entities.VoiceTranscript) string {
	text := ""
	hasUser := false
	for i := len(parts) - 1; i >= 0; i-- {
		part := "[" + parts[i].Role + "] " + parts[i].Text + "\n"
		if len(part) > 3500 {
			return ""
		}
		if len(text)+len(part) > 3500 {
			break
		}
		text = part + text
		if parts[i].Role == "user" && strings.TrimSpace(parts[i].Text) != "" {
			hasUser = true
		}
	}
	if !hasUser {
		return ""
	}
	return "Voice conversation transcript snapshot (may be partial or mistaken). Use the latest correction and clarify ambiguity before changes:\n" + text
}
func (c *VoiceSessions) Read(ctx context.Context, id string) (entities.VoiceRecord, error) {
	if len(id) != 38 || !strings.HasPrefix(id, "voice-") {
		return entities.VoiceRecord{}, ErrInvalidRecord
	}
	value, _, err := c.Agent.Repository.LoadVoice(ctx, id)
	c.mu.Lock()
	active := c.activeID == id
	c.mu.Unlock()
	if err == nil && !active && (value.Status == "starting" || value.Status == "active") {
		latest, _, readErr := c.Agent.Repository.LoadVoice(ctx, id)
		if readErr != nil {
			return value, readErr
		}
		value = latest
		if value.Status == "starting" || value.Status == "active" {
			value.Status = "interrupted"
			value.ErrorCode = "execution_not_active"
			value.Finalized = false
			value.UsageFinalConfirmed = false
		}
	}
	return value, err
}

func decodeVoiceInput(data []byte, value any) error {
	if len(data) > 32768 {
		return ErrInvalidRecord
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil || d.Decode(new(any)) != io.EOF {
		return ErrInvalidRecord
	}
	return nil
}

func (c *VoiceSessions) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.activeCancel != nil {
		c.activeStopReason = "settings_paused"
		c.activeCancel()
	}
}
