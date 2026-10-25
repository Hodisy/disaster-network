package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"yafa/backend/entities"
)

type RecordDailyStateInput struct {
	Date         *string `json:"date"`
	Expected     string  `json:"expected_revision"`
	Weather      *string `json:"weather"`
	Mood         *int    `json:"mood"`
	Energy       *int    `json:"energy"`
	MentalLoad   *int    `json:"mental_load"`
	WorkPressure *int    `json:"work_pressure"`
	SourceQuote  string  `json:"source_quote"`
}

func stateDate(raw *string, now time.Time) (string, error) {
	zone, err := parisZone()
	if err != nil {
		return "", err
	}
	local := now.In(zone)
	date := local.Format("2006-01-02")
	if raw != nil {
		switch *raw {
		case "today":
		case "yesterday":
			date = local.AddDate(0, 0, -1).Format("2006-01-02")
		default:
			date = *raw
		}
	}
	_, err = agendaDate(date)
	return date, err
}

func validWeather(value string) bool {
	return value == "sunny" || value == "mixed" || value == "cloudy" || value == "stormy"
}
func validGauge(name string, value int) bool {
	return (name == "mood" || name == "energy" || name == "mental_load" || name == "work_pressure") && value >= 1 && value <= 5
}
func validateDailyState(value entities.DailyState) error {
	if _, err := revision(value.Revision); err != nil {
		return ErrMemoryUnavailable
	}
	if value.Weather != nil && !validWeather(value.Weather.Value) {
		return ErrMemoryUnavailable
	}
	for name, gauge := range value.Gauges {
		if !validGauge(name, gauge.Value) {
			return ErrMemoryUnavailable
		}
	}
	return nil
}
func dailyStateView(value entities.DailyState) map[string]any {
	return map[string]any{"date": value.Date, "revision": value.Revision, "weather": value.Weather, "gauges": value.Gauges, "status": "recorded", "interpretation": "Subjective, revisable observations. Missing gauges are unknown, not zero. Not a diagnosis or actual weather."}
}

func (c *AgentHarness) ReadDailyState(ctx context.Context, raw *string, now time.Time) (map[string]any, error) {
	date, err := stateDate(raw, now)
	if err != nil {
		return nil, err
	}
	value, _, err := c.Repository.LoadDailyState(ctx, date)
	if errors.Is(err, ErrRecordNotFound) {
		return map[string]any{"date": date, "revision": "0", "status": "not_recorded", "weather": nil, "gauges": map[string]any{}, "interpretation": "No observation recorded; do not infer a score."}, nil
	}
	if err != nil {
		return nil, err
	}
	if err = validateDailyState(value); err != nil {
		return nil, err
	}
	return dailyStateView(value), nil
}

func (c *AgentHarness) RecordDailyState(ctx context.Context, runID, threadID string, in RecordDailyStateInput, fingerprint string) (json.RawMessage, error) {
	now := c.Workspace.Now()
	date, err := stateDate(in.Date, now)
	if err != nil {
		return nil, err
	}
	today, err := stateDate(nil, now)
	if err != nil {
		return nil, err
	}
	if date > today || !validText(in.SourceQuote, 300) {
		return nil, ErrInvalidRecord
	}
	if in.Weather != nil && !validWeather(*in.Weather) {
		return nil, ErrInvalidRecord
	}
	gauges := map[string]*int{"mood": in.Mood, "energy": in.Energy, "mental_load": in.MentalLoad, "work_pressure": in.WorkPressure}
	changed := in.Weather != nil
	for name, value := range gauges {
		if value != nil {
			if !validGauge(name, *value) {
				return nil, ErrInvalidRecord
			}
			changed = true
		}
	}
	if !changed {
		return nil, ErrInvalidRecord
	}
	run, _, err := c.Repository.LoadRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.ThreadID != threadID {
		return nil, ErrInvalidRecord
	}
	thread, err := c.Thread(ctx, threadID)
	if err != nil {
		return nil, err
	}
	quoted := false
	for _, message := range thread.Messages {
		if message.Role != "user" || message.RunID != runID {
			continue
		}
		if strings.HasPrefix(run.EventID, "voice-") {
			// Live requests include speaker-tagged transcripts. Accept only a quote
			// starting a user segment, never an assistant segment or older thread.
			quoted = strings.Contains(message.Text, "[user] "+in.SourceQuote)
		} else {
			quoted = strings.Contains(message.Text, in.SourceQuote)
		}
		if quoted {
			break
		}
	}
	if !quoted {
		return nil, ErrInvalidRecord
	}
	actionID := stableID("action", runID+":"+fingerprint)
	for attempt := 0; attempt < 3; attempt++ {
		workspace, _, err := c.Workspace.Read(ctx)
		if err != nil {
			return nil, err
		}
		if workspace.Lease == nil || workspace.Lease.RunID != runID || !c.Workspace.Now().Before(workspace.Lease.ExpiresAt) {
			return nil, ErrAgentBusy
		}
		value, version, err := c.Repository.LoadDailyState(ctx, date)
		if errors.Is(err, ErrRecordNotFound) {
			value = entities.DailyState{Date: date, Revision: "0", Gauges: map[string]entities.DailyGauge{}, Actions: map[string]entities.ActionReceipt{}}
			version = "0"
			err = nil
		} else if err == nil {
			err = validateDailyState(value)
		}
		if err != nil {
			return nil, err
		}
		if receipt, ok := value.Actions[actionID]; ok {
			if receipt.Fingerprint != fingerprint {
				return nil, ErrVersionConflict
			}
			return receipt.Result, nil
		}
		if value.Revision != in.Expected {
			return nil, ErrVersionConflict
		}
		if len(value.Actions) >= 64 {
			return nil, ErrAgentCapacity
		}
		provenance := entities.StateProvenance{Quote: in.SourceQuote, ThreadID: threadID, RunID: runID, UpdatedAt: now.UTC(), IsEstimate: true}
		if in.Weather != nil {
			value.Weather = &entities.DailyWeather{Value: *in.Weather, Provenance: provenance}
		}
		for name, score := range gauges {
			if score != nil {
				value.Gauges[name] = entities.DailyGauge{Value: *score, Provenance: provenance}
			}
		}
		if value.Revision == "0" {
			value.Revision = "1"
		} else {
			value.Revision, err = revision(value.Revision)
			if err != nil {
				return nil, err
			}
		}
		updates := map[string]int{}
		for name, score := range gauges {
			if score != nil {
				updates[name] = *score
			}
		}
		result := encoded(map[string]any{"date": date, "revision": value.Revision, "updated_gauges": updates, "updated_weather": in.Weather, "is_estimate": true, "source_quote": in.SourceQuote})
		value.Actions[actionID] = entities.ActionReceipt{Fingerprint: fingerprint, Result: result}
		_, err = c.Repository.SaveDailyState(ctx, version, value)
		if !errors.Is(err, ErrVersionConflict) {
			return result, err
		}
	}
	return nil, ErrVersionConflict
}

func workProgress(agenda entities.Agenda, workspace entities.Workspace) map[string]any {
	planned, complete := 0, 0
	for _, task := range workspace.Tasks {
		if task.ProjectID == nil || workspace.Projects[*task.ProjectID].Kind != "professional" {
			continue
		}
		date := ""
		if task.ScheduledDate != nil {
			date = *task.ScheduledDate
		} else if task.DueAt != nil {
			zone, err := parisZone()
			if err == nil {
				date = task.DueAt.In(zone).Format("2006-01-02")
			}
		}
		if date != agenda.StartDate {
			continue
		}
		planned++
		if task.Completed {
			complete++
		}
	}
	return map[string]any{"known_work_tasks_for_day": planned, "currently_checked": complete, "basis": "Current stored checkbox state for dated professional tasks, not hours worked or historical completions. Not a mood score."}
}
