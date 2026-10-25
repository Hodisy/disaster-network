package entities

import "time"

type DeviceDisplayCommand struct {
	Version    string `json:"version"`
	Expression string `json:"expression"`
	Awake      bool   `json:"awake"`
}

type DeviceReport struct {
	DeviceID       string `json:"device_id"`
	Firmware       string `json:"firmware"`
	BootID         string `json:"boot_id"`
	UptimeMS       uint64 `json:"uptime_ms"`
	Display        bool   `json:"display"`
	Touch          bool   `json:"touch"`
	Microphone     bool   `json:"microphone"`
	Speaker        bool   `json:"speaker"`
	Expression     string `json:"expression"`
	Awake          bool   `json:"awake"`
	Audio          string `json:"audio"`
	AppliedVersion string `json:"applied_version"`
}

// Link records contain actual reports and commands, never Wi-Fi/API credentials.
type DeviceLink struct {
	DeviceID   string               `json:"device_id"`
	LastSeenAt time.Time            `json:"last_seen_at"`
	Report     DeviceReport         `json:"report"`
	Desired    DeviceDisplayCommand `json:"desired"`
}

type DeviceExchange struct {
	Protocol int                  `json:"protocol"`
	DeviceID string               `json:"device_id"`
	Desired  DeviceDisplayCommand `json:"desired"`
}
