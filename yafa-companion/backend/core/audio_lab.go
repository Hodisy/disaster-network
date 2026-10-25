package core

import (
	"context"
	"encoding/binary"
	"errors"
	"yafa/backend/entities"
	"yafa/backend/services"
)

var ErrAudioLabUnavailable = errors.New("audio lab provider unavailable")
var ErrInvalidLabAudio = errors.New("invalid lab audio")

type AudioLab struct{ Service services.AudioLab }

func (c AudioLab) Prepare(ctx context.Context, mp3 []byte) (entities.LabSpeech, error) {
	if len(mp3) < 32 || len(mp3) > 2097152 {
		return entities.LabSpeech{}, ErrInvalidLabAudio
	}
	if c.Service == nil {
		return entities.LabSpeech{}, ErrAudioLabUnavailable
	}
	return c.Service.Prepare(ctx, mp3)
}
func validLabWAV(data []byte) bool {
	if len(data) < 44 || len(data) > 320080 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" || uint64(binary.LittleEndian.Uint32(data[4:8]))+8 != uint64(len(data)) {
		return false
	}
	format, samples := false, false
	for offset := 12; offset+8 <= len(data); {
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		start := offset + 8
		if size < 0 || size > len(data)-start {
			return false
		}
		switch string(data[offset : offset+4]) {
		case "fmt ":
			if size < 16 || format {
				return false
			}
			f := data[start : start+16]
			format = binary.LittleEndian.Uint16(f[:2]) == 1 && binary.LittleEndian.Uint16(f[2:4]) == 1 && binary.LittleEndian.Uint32(f[4:8]) == 16000 && binary.LittleEndian.Uint16(f[12:14]) == 2 && binary.LittleEndian.Uint16(f[14:16]) == 16
		case "data":
			if samples || size < 3200 || size > 320000 || size%2 != 0 {
				return false
			}
			samples = true
		}
		offset = start + size + (size & 1)
	}
	return format && samples
}
func (c AudioLab) Transcribe(ctx context.Context, wav []byte, consent bool) (entities.LabTranscript, error) {
	if !consent {
		return entities.LabTranscript{}, ErrConsentRequired
	}
	if !validLabWAV(wav) {
		return entities.LabTranscript{}, ErrInvalidLabAudio
	}
	if c.Service == nil {
		return entities.LabTranscript{}, ErrAudioLabUnavailable
	}
	return c.Service.Transcribe(ctx, wav)
}
