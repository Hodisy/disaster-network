package services

import (
	"context"
	"yafa/backend/entities"
)

type AudioLab interface {
	Prepare(context.Context, []byte) (entities.LabSpeech, error)
	Transcribe(context.Context, []byte) (entities.LabTranscript, error)
}
