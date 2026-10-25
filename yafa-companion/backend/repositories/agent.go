package repositories

import (
	"context"
	"encoding/json"
	"yafa/backend/entities"
)

type AgentRepository interface {
	LoadWorkspace(context.Context) (entities.Workspace, string, error)
	SaveWorkspace(context.Context, string, entities.Workspace) (string, error)
	LoadThread(context.Context, string) (entities.Thread, string, error)
	SaveThread(context.Context, string, entities.Thread) (string, error)
	LoadRun(context.Context, string) (entities.AgentRun, string, error)
	SaveRun(context.Context, string, entities.AgentRun) (string, error)
	SaveArtifact(context.Context, string, json.RawMessage) error
	LoadArtifact(context.Context, string) (json.RawMessage, error)
	LoadVoice(context.Context, string) (entities.VoiceRecord, string, error)
	SaveVoice(context.Context, string, entities.VoiceRecord) (string, error)
	LoadDayActivity(context.Context, string) (entities.DayActivity, string, error)
	SaveDayActivity(context.Context, string, entities.DayActivity) (string, error)
	LoadDailyState(context.Context, string) (entities.DailyState, string, error)
	SaveDailyState(context.Context, string, entities.DailyState) (string, error)
}
