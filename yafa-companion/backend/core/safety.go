package core

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

var ErrSafetyBlocked = errors.New("AI safety limit or pause")

const budgetLedgerKey = "agent-budget-rolling-v1"

type SafetyQuota struct {
	CodexJobs     int     `json:"codex_jobs"`
	Runs          int     `json:"runs"`
	Tokens        int64   `json:"text_tokens"`
	VoiceSessions int     `json:"voice_sessions"`
	VoiceSeconds  float64 `json:"voice_seconds"`
	Hosted        int     `json:"hosted_jobs"`
	Images        int     `json:"images"`
	KnownUSD      float64 `json:"known_cost_stop_usd"`
}
type SafetyLimits struct {
	Enabled   bool        `json:"paid_ai_enabled"`
	Window5h  SafetyQuota `json:"window_5h"`
	Window24h SafetyQuota `json:"window_24h"`
}
type SafetySettings struct {
	Paused    bool        `json:"paused"`
	Window5h  SafetyQuota `json:"window_5h"`
	Window24h SafetyQuota `json:"window_24h"`
}
type BudgetEntry struct {
	Kind            string    `json:"kind"`
	CreatedAt       time.Time `json:"created_at"`
	Tokens          int64     `json:"tokens_used_or_reserved"`
	Seconds         float64   `json:"seconds_used_or_reserved"`
	Settled         bool      `json:"settled"`
	UsageConfirmed  bool      `json:"usage_confirmed"`
	KnownUSD        float64   `json:"known_estimated_usd"`
	CostUnconfirmed bool      `json:"cost_unconfirmed"`
	Status          string    `json:"status"`
	ErrorCode       string    `json:"error_code,omitempty"`
}
type BudgetDay struct {
	Entries map[string]BudgetEntry `json:"entries"`
}
type SafetyCore struct {
	Repository repositories.MemoryRepository
	Limits     SafetyLimits
}

func validQuota(q SafetyQuota) bool {
	return q.CodexJobs >= 0 && q.CodexJobs <= 20 && q.Runs >= 1 && q.Runs <= 200 && q.Tokens >= 1024 && q.Tokens <= 10000000 && q.VoiceSessions >= 1 && q.VoiceSessions <= 100 && q.VoiceSeconds >= 30 && q.VoiceSeconds <= 86400 && q.Hosted >= 1 && q.Hosted <= 100 && q.Images >= 1 && q.Images <= q.Hosted && q.KnownUSD > 0 && q.KnownUSD <= 1000 && !math.IsNaN(q.KnownUSD) && !math.IsInf(q.KnownUSD, 0) && !math.IsNaN(q.VoiceSeconds) && !math.IsInf(q.VoiceSeconds, 0)
}
func validSafetySettings(s SafetySettings) bool {
	a, b := s.Window5h, s.Window24h
	return validQuota(a) && validQuota(b) && a.CodexJobs <= b.CodexJobs && a.Runs <= b.Runs && a.Tokens <= b.Tokens && a.VoiceSessions <= b.VoiceSessions && a.VoiceSeconds <= b.VoiceSeconds && a.Hosted <= b.Hosted && a.Images <= b.Images && a.KnownUSD <= b.KnownUSD
}
func (c *SafetyCore) Settings(ctx context.Context) (SafetySettings, string, error) {
	value := SafetySettings{Window5h: c.Limits.Window5h, Window24h: c.Limits.Window24h}
	row, err := c.Repository.Read(ctx, "agent-safety-settings")
	if errors.Is(err, ErrRecordNotFound) {
		return value, "0", nil
	}
	if err != nil {
		return value, "", err
	}
	if json.Unmarshal(row.Value, &value) != nil || !validSafetySettings(value) {
		return value, "", ErrMemoryUnavailable
	}
	return value, row.Version, nil
}
func (c *SafetyCore) Check(ctx context.Context) error {
	if c == nil || c.Repository == nil || !c.Limits.Enabled {
		return ErrSafetyBlocked
	}
	settings, _, err := c.Settings(ctx)
	if err != nil {
		return err
	}
	if settings.Paused {
		return ErrSafetyBlocked
	}
	return nil
}
func (c *SafetyCore) Update(ctx context.Context, expected string, settings SafetySettings) error {
	if !validSafetySettings(settings) {
		return ErrInvalidRecord
	}
	_, err := c.Repository.Write(ctx, "agent-safety-settings", expected, encoded(settings))
	return err
}
func (c *SafetyCore) readLedger(ctx context.Context) (BudgetDay, string, error) {
	day := BudgetDay{Entries: map[string]BudgetEntry{}}
	row, err := c.Repository.Read(ctx, budgetLedgerKey)
	if errors.Is(err, ErrRecordNotFound) {
		return day, "0", nil
	}
	if err != nil {
		return day, "", err
	}
	if json.Unmarshal(row.Value, &day) != nil || day.Entries == nil || len(day.Entries) > 512 {
		return day, "", ErrMemoryUnavailable
	}
	for _, e := range day.Entries {
		if e.CreatedAt.IsZero() || e.Tokens < 0 || e.Seconds < 0 || math.IsNaN(e.Seconds) || math.IsInf(e.Seconds, 0) || e.KnownUSD < 0 || math.IsNaN(e.KnownUSD) || math.IsInf(e.KnownUSD, 0) {
			return day, "", ErrMemoryUnavailable
		}
	}
	return day, row.Version, nil
}
func quotaUsage(day BudgetDay, now time.Time, width time.Duration) (SafetyQuota, int64, float64, bool, *time.Time) {
	q := SafetyQuota{}
	var reservedTokens int64
	reservedSeconds := 0.0
	unknown := false
	var next *time.Time
	for _, e := range day.Entries {
		if !e.CreatedAt.After(now.Add(-width)) {
			continue
		}
		switch e.Kind {
		case "codex":
			q.CodexJobs++
		case "run":
			q.Runs++
		case "voice":
			q.VoiceSessions++
		case "image":
			q.Images++
			q.Hosted++
		case "analysis":
			q.Hosted++
		}
		q.Tokens += e.Tokens
		q.VoiceSeconds += e.Seconds
		q.KnownUSD += e.KnownUSD
		if !e.UsageConfirmed {
			reservedTokens += e.Tokens
			reservedSeconds += e.Seconds
		}
		if e.CostUnconfirmed {
			unknown = true
		}
		expires := e.CreatedAt.Add(width)
		if next == nil || expires.Before(*next) {
			value := expires
			next = &value
		}
	}
	return q, reservedTokens, reservedSeconds, unknown, next
}
func quotaFits(used, limit SafetyQuota) bool {
	return used.CodexJobs <= limit.CodexJobs && used.Runs <= limit.Runs && used.Tokens <= limit.Tokens && used.VoiceSessions <= limit.VoiceSessions && used.VoiceSeconds <= limit.VoiceSeconds && used.Hosted <= limit.Hosted && used.Images <= limit.Images && used.KnownUSD < limit.KnownUSD
}
func quotaRatio(q, limit SafetyQuota) float64 {
	ratio := 0.0
	if limit.CodexJobs > 0 {
		ratio = float64(q.CodexJobs) / float64(limit.CodexJobs)
	}
	return math.Max(ratio, baseQuotaRatio(q, limit))
}
func baseQuotaRatio(q, limit SafetyQuota) float64 {
	return math.Max(float64(q.Runs)/float64(limit.Runs), math.Max(float64(q.Tokens)/float64(limit.Tokens), math.Max(float64(q.VoiceSessions)/float64(limit.VoiceSessions), math.Max(q.VoiceSeconds/limit.VoiceSeconds, math.Max(float64(q.Hosted)/float64(limit.Hosted), math.Max(float64(q.Images)/float64(limit.Images), q.KnownUSD/limit.KnownUSD))))))
}
func (c *SafetyCore) Status(ctx context.Context) (map[string]any, error) {
	settings, revision, err := c.Settings(ctx)
	if err != nil {
		return nil, err
	}
	ledger, _, err := c.readLedger(ctx)
	if err != nil {
		return nil, err
	}
	windows := map[string]any{}
	now := time.Now().UTC()
	for name, width := range map[string]time.Duration{"5h": 5 * time.Hour, "24h": 24 * time.Hour} {
		limit := settings.Window5h
		if name == "24h" {
			limit = settings.Window24h
		}
		used, tokens, seconds, unknown, next := quotaUsage(ledger, now, width)
		state := "available"
		ratio := quotaRatio(used, limit)
		if ratio >= 1 {
			state = "at_limit"
		} else if ratio >= 0.8 {
			state = "approaching_limit"
		}
		windows[name] = map[string]any{"limits": limit, "used_or_reserved": used, "reserved_tokens": tokens, "reserved_voice_seconds": seconds, "cost_unconfirmed": unknown, "next_entry_expires_at": next, "state": state}
	}
	type activity struct {
		ID string `json:"id"`
		BudgetEntry
	}
	recent := []activity{}
	for id, entry := range ledger.Entries {
		recent = append(recent, activity{ID: id, BudgetEntry: entry})
	}
	sort.Slice(recent, func(i, j int) bool { return recent[i].CreatedAt.After(recent[j].CreatedAt) })
	if len(recent) > 20 {
		recent = recent[:20]
	}
	return map[string]any{"recent_activity": recent, "paid_ai_enabled": c.Limits.Enabled, "settings": settings, "revision": revision, "windows": windows, "as_of": now, "currency": "USD", "cost_note": "Known estimated costs only. Missing usage/prices are not zero. The currency stop threshold cannot cap an in-flight invoice; unit quotas reserve capacity before calls. Windows use request start time."}, nil
}
func (c *SafetyCore) Reserve(ctx context.Context, id, kind string, tokens int64, seconds float64) (string, error) {
	if err := c.Check(ctx); err != nil {
		return "", err
	}
	if !recordKey.MatchString(id) || tokens < 0 || seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) || (kind != "run" && kind != "voice" && kind != "image" && kind != "analysis" && kind != "codex") {
		return "", ErrInvalidRecord
	}
	for attempt := 0; attempt < 4; attempt++ {
		settings, _, err := c.Settings(ctx)
		if err != nil {
			return "", err
		}
		if settings.Paused {
			return "", ErrSafetyBlocked
		}
		day, version, err := c.readLedger(ctx)
		if err != nil {
			return "", err
		}
		if _, exists := day.Entries[id]; exists {
			return "", ErrSafetyBlocked
		}
		now := time.Now().UTC()
		for key, e := range day.Entries {
			if e.CreatedAt.Before(now.Add(-24 * time.Hour)) {
				delete(day.Entries, key)
			}
		}
		if len(day.Entries) >= 512 {
			return "", ErrSafetyBlocked
		}
		day.Entries[id] = BudgetEntry{Kind: kind, CreatedAt: now, Tokens: tokens, Seconds: seconds, CostUnconfirmed: true, Status: "reserved"}
		short, _, _, _, _ := quotaUsage(day, now, 5*time.Hour)
		long, _, _, _, _ := quotaUsage(day, now, 24*time.Hour)
		if !quotaFits(short, settings.Window5h) || !quotaFits(long, settings.Window24h) {
			return "", ErrSafetyBlocked
		}
		_, err = c.Repository.Write(ctx, budgetLedgerKey, version, encoded(day))
		if errors.Is(err, ErrVersionConflict) {
			continue
		}
		return budgetLedgerKey, err
	}
	return "", ErrVersionConflict
}

// Unknown/interrupted usage retains its reservation until it leaves the window.
func (c *SafetyCore) Settle(ctx context.Context, key, id, status string, tokens int64, seconds, known float64, usageConfirmed, costConfirmed bool, errorCode ...string) error {
	if c == nil {
		return nil
	}
	if key != budgetLedgerKey || tokens < 0 || seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) || known < 0 || math.IsNaN(known) || math.IsInf(known, 0) {
		return ErrInvalidRecord
	}
	for attempt := 0; attempt < 4; attempt++ {
		day, version, err := c.readLedger(ctx)
		if err != nil {
			return err
		}
		e, ok := day.Entries[id]
		if !ok {
			return ErrRecordNotFound
		}
		if e.Settled {
			return nil
		}
		if usageConfirmed {
			e.Tokens = tokens
			e.Seconds = seconds
		}
		e.UsageConfirmed = usageConfirmed
		e.KnownUSD = known
		e.CostUnconfirmed = !costConfirmed
		e.Settled = true
		e.Status = status
		if len(errorCode) > 0 {
			e.ErrorCode = clipText(errorCode[0], 80)
		}
		day.Entries[id] = e
		_, err = c.Repository.Write(ctx, budgetLedgerKey, version, encoded(day))
		if errors.Is(err, ErrVersionConflict) {
			continue
		}
		return err
	}
	return ErrVersionConflict
}

func refreshUsageCost(u *entities.Usage) {
	if u.Incomplete || u.AdditionalCostsUnpriced {
		u.EstimatedCostUSD = nil
		return
	}
	value := u.KnownCostUSD
	u.EstimatedCostUSD = &value
}
