package services

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os/exec"
	"yafa/backend/core"
	"yafa/backend/entities"
)

type OpenAIAudioLab struct {
	Key                     string
	Client                  *http.Client
	TranscriptModel, FFmpeg string
}

func (s OpenAIAudioLab) request(ctx context.Context, path, contentType string, body []byte, limit int64) ([]byte, error) {
	if s.Key == "" || s.Client == nil {
		return nil, core.ErrAudioLabUnavailable
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/"+path, bytes.NewReader(body))
	if err != nil {
		return nil, core.ErrAudioLabUnavailable
	}
	r.Header.Set("Authorization", "Bearer "+s.Key)
	r.Header.Set("Content-Type", contentType)
	response, err := s.Client.Do(r)
	if err != nil {
		return nil, core.ErrAudioLabUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%w (provider HTTP %d)", core.ErrAudioLabUnavailable, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, core.ErrAudioLabUnavailable
	}
	return data, nil
}

// Prepare only decodes a user-supplied local asset. It makes no provider call.
func (s OpenAIAudioLab) Prepare(ctx context.Context, mp3 []byte) (entities.LabSpeech, error) {
	format := "mp3"
	if len(mp3) >= 12 && string(mp3[:4]) == "RIFF" && string(mp3[8:12]) == "WAVE" {
		format = "wav"
	}
	program := s.FFmpeg
	if program == "" {
		program = "ffmpeg"
	}
	command := exec.CommandContext(ctx, program, "-hide_banner", "-loglevel", "error", "-protocol_whitelist", "pipe", "-f", format, "-i", "pipe:0", "-t", "15", "-vn", "-ac", "1", "-ar", "16000", "-f", "s16le", "pipe:1")
	command.Stdin = bytes.NewReader(mp3)
	pcm, err := command.Output()
	if err != nil || len(pcm) < 3200 || len(pcm) > 480000 || len(pcm)%2 != 0 {
		return entities.LabSpeech{}, core.ErrAudioLabUnavailable
	}
	// Normalize the asset on the PC; the device needs no whole-clip peak scan.
	peak := 0
	for i := 0; i < len(pcm); i += 2 {
		v := int(int16(binary.LittleEndian.Uint16(pcm[i : i+2])))
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	if peak > 0 {
		gain := 4000.0 / float64(peak)
		if gain > 8 {
			gain = 8
		}
		for i := 0; i < len(pcm); i += 2 {
			v := int16(binary.LittleEndian.Uint16(pcm[i : i+2]))
			binary.LittleEndian.PutUint16(pcm[i:i+2], uint16(int16(float64(v)*gain)))
		}
	}
	return entities.LabSpeech{Audio: mp3, Format: format, PCM: pcm, SampleRate: 16000, TruncatedAtSeconds: 15}, nil
}
func (s OpenAIAudioLab) Transcribe(ctx context.Context, wav []byte) (entities.LabTranscript, error) {
	var buffer bytes.Buffer
	form := multipart.NewWriter(&buffer)
	file, err := form.CreateFormFile("file", "device-recording.wav")
	if err != nil {
		return entities.LabTranscript{}, core.ErrAudioLabUnavailable
	}
	_, _ = file.Write(wav)
	model := s.TranscriptModel
	if model == "" {
		model = "gpt-transcribe"
	}
	_ = form.WriteField("model", model)
	_ = form.Close()
	data, err := s.request(ctx, "audio/transcriptions", form.FormDataContentType(), buffer.Bytes(), 131072)
	if err != nil {
		return entities.LabTranscript{}, err
	}
	var transcript struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(data, &transcript) != nil {
		return entities.LabTranscript{}, core.ErrAudioLabUnavailable
	}
	return entities.LabTranscript{Text: transcript.Text, Model: model}, nil
}
