package core

import (
	"context"
	"errors"
	"time"

	"yafa/backend/entities"
	"yafa/backend/repositories"
	"yafa/backend/services"
)

var (
	ErrDeviceAccessDenied = errors.New("device access denied")
	ErrDeviceOffline      = errors.New("device offline")
	ErrConsentRequired    = errors.New("explicit audio start required")
	ErrAudioUnsupported   = errors.New("device has no microphone and speaker path")
	ErrAudioUnavailable   = errors.New("audio transport unavailable")
	ErrAudioFormat        = errors.New("audio format unsupported")
	ErrNotConfigured      = errors.New("device core not configured")
)

// Devices contains use cases only. HTTP, SQL and external-provider code belong
// in connectors; clocks are supplied at composition time rather than hidden here.
type Devices struct {
	Repository repositories.DeviceRepository
	Audio      services.AudioTransport
	Now        func() time.Time
}

func (c Devices) ReadDevice(ctx context.Context, ownerID, deviceID string) (entities.Device, error) {
	if c.Repository == nil {
		return entities.Device{}, ErrNotConfigured
	}
	if ownerID == "" || deviceID == "" {
		return entities.Device{}, ErrDeviceAccessDenied
	}
	device, err := c.Repository.GetDevice(ctx, deviceID)
	if err != nil {
		return entities.Device{}, err
	}
	if device.ID != deviceID || device.OwnerID != ownerID || device.Disabled {
		return entities.Device{}, ErrDeviceAccessDenied
	}
	return device, nil
}

// PrepareVoice validates a future bidirectional voice path. It performs no
// microphone capture, session creation, playback or external AI call.
// The display-only Feather must fail with ErrAudioUnsupported.
func (c Devices) PrepareVoice(ctx context.Context, ownerID, deviceID string, format entities.AudioFormat, explicitlyStarted bool) (entities.VoicePlan, error) {
	if !explicitlyStarted {
		return entities.VoicePlan{}, ErrConsentRequired
	}
	device, err := c.ReadDevice(ctx, ownerID, deviceID)
	if err != nil {
		return entities.VoicePlan{}, err
	}
	if !device.Capabilities.Microphone || !device.Capabilities.Speaker {
		return entities.VoicePlan{}, ErrAudioUnsupported
	}
	if c.Now == nil {
		return entities.VoicePlan{}, ErrNotConfigured
	}
	age := c.Now().Sub(device.LastSeenAt)
	if device.LastSeenAt.IsZero() || age < 0 || age > 90*time.Second {
		return entities.VoicePlan{}, ErrDeviceOffline
	}
	if format.Encoding != "pcm_s16le" || format.Channels != 1 || format.SampleRate < 8000 || format.SampleRate > 48000 {
		return entities.VoicePlan{}, ErrAudioFormat
	}
	if c.Audio == nil {
		return entities.VoicePlan{}, ErrAudioUnavailable
	}
	available, err := c.Audio.Available(ctx)
	if err != nil {
		return entities.VoicePlan{}, err
	}
	if !available {
		return entities.VoicePlan{}, ErrAudioUnavailable
	}
	supported, err := c.Audio.SupportsFormat(ctx, format)
	if err != nil {
		return entities.VoicePlan{}, err
	}
	if !supported {
		return entities.VoicePlan{}, ErrAudioFormat
	}
	return entities.VoicePlan{DeviceID: device.ID, Format: format}, nil
}
