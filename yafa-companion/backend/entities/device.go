package entities

import "time"

// Capabilities describe the actual paired hardware, not the intended future device.
type DeviceCapabilities struct {
	Display    bool
	Microphone bool
	Speaker    bool
}

type Device struct {
	ID           string
	OwnerID      string
	Capabilities DeviceCapabilities
	LastSeenAt   time.Time
	Disabled     bool
	Version      uint64
}

// Presentation is a desired display state. It does not establish device delivery.
type Presentation struct {
	Expression string
	Text       string
	TaskID     string
	Brightness uint8
	Version    uint64
}

type AudioFormat struct {
	Encoding   string
	SampleRate int
	Channels   int
}

// VoicePlan is only a validated configuration. It is not an opened audio stream.
type VoicePlan struct {
	DeviceID string
	Format   AudioFormat
}
