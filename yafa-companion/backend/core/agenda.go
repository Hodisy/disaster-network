package core

import (
	"context"
	"errors"
	"sort"
	"time"
	"yafa/backend/entities"
)

func parisZone() (*time.Location, error) { return time.LoadLocation("Europe/Paris") }

func agendaDate(value string) (time.Time, error) {
	zone, err := parisZone()
	if err != nil {
		return time.Time{}, err
	}
	date, err := time.ParseInLocation("2006-01-02", value, zone)
	if err != nil || date.Year() < 1 || date.Format("2006-01-02") != value {
		return time.Time{}, ErrInvalidRecord
	}
	return date, nil
}

func validSchedule(date, clock *string) bool {
	if clock != nil && date == nil {
		return false
	}
	if date != nil {
		if _, err := agendaDate(*date); err != nil {
			return false
		}
	}
	if clock != nil {
		value, err := time.Parse("15:04", *clock)
		if err != nil || value.Format("15:04") != *clock {
			return false
		}
	}
	return true
}

type AgendaInput struct {
	Date             *string `json:"date"`
	Owner            string  `json:"owner"`
	IncludeCompleted bool    `json:"include_completed"`
	Offset           int     `json:"offset"`
	Limit            int     `json:"limit"`
}

// Week ranges use local calendar days, not multiples of 24 hours across DST.
func (c *AgentHarness) ReadAgenda(ctx context.Context, in AgendaInput, week bool, now time.Time) (entities.Agenda, error) {
	zone, err := parisZone()
	if err != nil {
		return entities.Agenda{}, err
	}
	local := now.In(zone)
	date := local.Format("2006-01-02")
	if in.Date != nil {
		date = *in.Date
	}
	switch date {
	case "today":
		date = local.Format("2006-01-02")
	case "yesterday":
		date = local.AddDate(0, 0, -1).Format("2006-01-02")
	case "tomorrow":
		date = local.AddDate(0, 0, 1).Format("2006-01-02")
	case "this_week", "last_week", "next_week":
		if !week {
			return entities.Agenda{}, ErrInvalidRecord
		}
		delta := 0
		if date == "last_week" {
			delta = -7
		}
		if date == "next_week" {
			delta = 7
		}
		date = local.AddDate(0, 0, delta).Format("2006-01-02")
	}
	start, err := agendaDate(date)
	if err != nil {
		return entities.Agenda{}, err
	}
	if week {
		start = start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7))
	}
	days := 1
	if week {
		days = 7
	}
	return c.readAgendaPeriod(ctx, in, start, days)
}

func (c *AgentHarness) readAgendaPeriod(ctx context.Context, in AgendaInput, start time.Time, days int) (entities.Agenda, error) {
	if days < 1 || days > 31 || start.Year() < 1 || start.AddDate(0, 0, days-1).Year() > 9999 {
		return entities.Agenda{}, ErrInvalidRecord
	}
	workspace, _, err := c.Workspace.Read(ctx)
	if err != nil {
		return entities.Agenda{}, err
	}
	return c.agenda(ctx, workspace, in, start, days)
}

// Range endpoints are inclusive local dates. Limit work before reading any data.
func (c *AgentHarness) ReadAgendaRange(ctx context.Context, in AgendaInput, startDate, endDate string) (entities.Agenda, error) {
	start, err := agendaDate(startDate)
	if err != nil {
		return entities.Agenda{}, err
	}
	end, err := agendaDate(endDate)
	if err != nil {
		return entities.Agenda{}, err
	}
	if end.Before(start) {
		return entities.Agenda{}, ErrInvalidRecord
	}
	days := 0
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		days++
		if days > 31 {
			return entities.Agenda{}, ErrInvalidRecord
		}
	}
	return c.readAgendaPeriod(ctx, in, start, days)
}

func (c *AgentHarness) ReadAgendaMonth(ctx context.Context, in AgendaInput, month *string, now time.Time) (entities.Agenda, error) {
	zone, err := parisZone()
	if err != nil {
		return entities.Agenda{}, err
	}
	local := now.In(zone)
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, zone)
	if month != nil {
		switch *month {
		case "this_month":
		case "last_month":
			start = start.AddDate(0, -1, 0)
		case "next_month":
			start = start.AddDate(0, 1, 0)
		default:
			start, err = time.ParseInLocation("2006-01", *month, zone)
			if err != nil || start.Year() < 1 || start.Format("2006-01") != *month {
				return entities.Agenda{}, ErrInvalidRecord
			}
		}
	}
	days := start.AddDate(0, 1, -1).Day()
	return c.readAgendaPeriod(ctx, in, start, days)
}

func (c *AgentHarness) agenda(ctx context.Context, workspace entities.Workspace, in AgendaInput, start time.Time, days int) (entities.Agenda, error) {
	result := entities.Agenda{StartDate: start.Format("2006-01-02"), EndDate: start.AddDate(0, 0, days-1).Format("2006-01-02"), Timezone: "Europe/Paris", Owner: in.Owner, Entries: []entities.AgendaEntry{}, Coverage: "Me: stored dated tasks; Yafa: recorded agent-run activity since indexing was introduced, not a backfill or a complete calendar. Undated tasks are separate."}
	if (in.Owner != "me" && in.Owner != "yafa" && in.Owner != "both") || days < 1 || days > 31 || in.Offset < 0 || in.Offset > 4096 || in.Limit < 1 || in.Limit > 20 {
		return result, ErrInvalidRecord
	}
	zone := start.Location()
	for _, task := range workspace.Tasks {
		if !validSchedule(task.ScheduledDate, task.ScheduledTime) {
			return result, ErrMemoryUnavailable
		}
		if task.ScheduledDate == nil && task.DueAt == nil {
			if in.IncludeCompleted || !task.Completed {
				result.UndatedTasks++
			}
			continue
		}
		if in.Owner == "yafa" || (!in.IncludeCompleted && task.Completed) {
			continue
		}
		date, source := "", "scheduled"
		var clock *string
		if task.ScheduledDate != nil {
			date, clock = *task.ScheduledDate, task.ScheduledTime
		} else {
			date, source = task.DueAt.In(zone).Format("2006-01-02"), "deadline"
			value := task.DueAt.In(zone).Format("15:04")
			clock = &value
		}
		if date >= result.StartDate && date <= result.EndDate {
			item := task
			result.Entries = append(result.Entries, entities.AgendaEntry{ID: task.ID, Owner: "me", Kind: "task", Date: date, Time: clock, DateSource: source, Task: &item})
		}
	}
	if in.Owner != "me" {
		for day := 0; day < days; day++ {
			date := start.AddDate(0, 0, day).Format("2006-01-02")
			activity, _, err := c.Repository.LoadDayActivity(ctx, date)
			if errors.Is(err, ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return result, err
			}
			for id, thread := range activity.Threads {
				if id != thread.ThreadID || !validAgentID(id, "thread") || !validOrigin(thread.Origin) || thread.UpdatedAt.In(zone).Format("2006-01-02") != date {
					return result, ErrMemoryUnavailable
				}
				item := thread
				clock := thread.UpdatedAt.In(zone).Format("15:04")
				result.Entries = append(result.Entries, entities.AgendaEntry{ID: date + ":" + id, Owner: "yafa", Kind: "thread_activity", Date: date, Time: &clock, DateSource: "recorded_activity", Thread: &item})
			}
		}
	}
	sort.Slice(result.Entries, func(i, j int) bool {
		a, b := result.Entries[i], result.Entries[j]
		if a.Date != b.Date {
			return a.Date < b.Date
		}
		if a.Time == nil && b.Time != nil {
			return false
		}
		if a.Time != nil && b.Time == nil {
			return true
		}
		if a.Time != nil && b.Time != nil && *a.Time != *b.Time {
			return *a.Time < *b.Time
		}
		return a.ID < b.ID
	})
	result.Total = len(result.Entries)
	from := min(in.Offset, result.Total)
	to := min(from+in.Limit, result.Total)
	if to < result.Total {
		next := to
		result.NextOffset = &next
	}
	result.Entries = result.Entries[from:to]
	return result, nil
}

// Only lightweight timeline rows enter instructions. Full notes stay behind tools.
// Bound the snapshot separately from the API page and keep pagination explicit.
func compactAgenda(agenda entities.Agenda, workspace entities.Workspace) map[string]any {
	entries := []map[string]any{}
	for _, entry := range agenda.Entries {
		row := map[string]any{"id": entry.ID, "owner": entry.Owner, "kind": entry.Kind, "date": entry.Date, "time": entry.Time, "date_source": entry.DateSource}
		if task := entry.Task; task != nil {
			row["title"], row["title_truncated"], row["completed"], row["revision"], row["project_id"] = clipText(task.Title, 100), len(task.Title) > 100, task.Completed, task.Revision, task.ProjectID
			if task.ProjectID != nil {
				if project, ok := workspace.Projects[*task.ProjectID]; ok {
					row["project_name"] = clipText(project.Name, 60)
				}
			}
		}
		if thread := entry.Thread; thread != nil {
			row["title"], row["thread_id"], row["origin"], row["last_run_id"] = thread.Title, thread.ThreadID, thread.Origin, thread.LastRunID
		}
		entries = append(entries, row)
	}
	value := map[string]any{"start_date": agenda.StartDate, "end_date": agenda.EndDate, "timezone": agenda.Timezone, "owner": agenda.Owner, "entries": entries, "total": agenda.Total, "next_offset": agenda.NextOffset, "undated_task_count": agenda.UndatedTasks, "coverage": agenda.Coverage}
	for len(encoded(value)) > 6000 && len(entries) > 0 {
		entries = entries[:len(entries)-1]
		value["entries"] = entries
		value["next_offset"] = len(entries)
	}
	return value
}

// Thread pointers are indexed, not copies of transcripts, prompts or private reasoning.
// One thread appears once per date; clients retrieve actual thread/run details.
func (c *AgentHarness) recordThreadActivity(ctx context.Context, thread entities.Thread, run entities.AgentRun) error {
	zone, err := parisZone()
	if err != nil {
		return err
	}
	at := run.CreatedAt
	date := at.In(zone).Format("2006-01-02")
	title := "Conversation with Yafa"
	for _, message := range thread.Messages {
		if message.Role == "user" {
			title = clipText(message.Text, 96)
			break
		}
	}
	for attempt := 0; attempt < 3; attempt++ {
		value, version, err := c.Repository.LoadDayActivity(ctx, date)
		if errors.Is(err, ErrRecordNotFound) {
			value = entities.DayActivity{Date: date, Threads: map[string]entities.ThreadActivity{}}
			version = "0"
			err = nil
		}
		if err != nil {
			return err
		}
		if old, ok := value.Threads[thread.ID]; ok && old.LastRunID == run.ID {
			return nil
		}
		if len(value.Threads) >= 128 {
			if _, ok := value.Threads[thread.ID]; !ok {
				return ErrAgentCapacity
			}
		}
		value.Threads[thread.ID] = entities.ThreadActivity{ThreadID: thread.ID, Title: title, Origin: thread.Origin, LastRunID: run.ID, UpdatedAt: at}
		_, err = c.Repository.SaveDayActivity(ctx, version, value)
		if !errors.Is(err, ErrVersionConflict) {
			return err
		}
	}
	return ErrVersionConflict
}
