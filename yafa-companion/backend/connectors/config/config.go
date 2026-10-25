package config

import (
	"errors"
	"math"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"yafa/backend/core"
	"yafa/backend/entities"
)

type Config struct {
	Listen                                                               string
	APIToken                                                             string
	MemoryURL                                                            string
	MemoryToken                                                          string
	DeviceID                                                             string
	DeviceToken                                                          string
	DeviceListen                                                         string
	OpenAIKey                                                            string
	OpenAIModel                                                          string
	ImageModel                                                           string
	LiveModel                                                            string
	LiveVoice                                                            string
	LiveMaxSeconds                                                       int
	AgentLimits                                                          entities.AgentLimits
	TokenPrices                                                          []float64
	Safety                                                               core.SafetyLimits
	LiveIdleSeconds                                                      int
	LivePrice                                                            *float64
	Features                                                             core.ToolFeatures
	CodexHost, CodexUser, CodexKeyFile, CodexKnownHostsFile, CodexRunner string
	CodexPort, CodexMaxSeconds                                           int
}

func Load() (Config, error) {
	c := Config{Listen: os.Getenv("LISTEN_ADDRESS"), APIToken: os.Getenv("YAFA_API_TOKEN"), MemoryURL: os.Getenv("MEMORY_URL"), MemoryToken: os.Getenv("YAFA_MEMORY_TOKEN")}
	c.DeviceID, c.DeviceToken, c.DeviceListen = os.Getenv("YAFA_DEVICE_ID"), os.Getenv("YAFA_DEVICE_TOKEN"), os.Getenv("YAFA_DEVICE_LISTEN")
	c.OpenAIKey = os.Getenv("OPENAI_API_KEY")
	if c.DeviceID != "" || c.DeviceToken != "" || c.DeviceListen != "" {
		if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,31}$`).MatchString(c.DeviceID) || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(c.DeviceToken) || c.DeviceToken == c.APIToken || c.DeviceToken == c.MemoryToken {
			return Config{}, errors.New("device configuration requires a valid ID and a separate 64-hex device token")
		}
		if c.DeviceListen == "" {
			c.DeviceListen = ":8081"
		}
	}
	if c.Listen == "" {
		c.Listen = ":8080"
	}
	if len(c.APIToken) < 32 || len(c.MemoryToken) < 32 {
		return Config{}, errors.New("YAFA_API_TOKEN and YAFA_MEMORY_TOKEN must each contain at least 32 characters")
	}
	u, err := url.Parse(c.MemoryURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return Config{}, errors.New("MEMORY_URL must be a service HTTP origin")
	}
	c.OpenAIKey = os.Getenv("OPENAI_API_KEY")
	c.OpenAIModel = os.Getenv("OPENAI_MODEL")
	c.ImageModel = os.Getenv("OPENAI_IMAGE_MODEL")
	c.LiveModel = os.Getenv("OPENAI_LIVE_MODEL")
	if c.LiveModel == "" {
		c.LiveModel = "gpt-live-1"
	}
	c.LiveVoice = os.Getenv("OPENAI_LIVE_VOICE")
	if c.LiveVoice == "" {
		c.LiveVoice = "marin"
	}
	if strings.ContainsAny(c.OpenAIKey, "\r\n") || len(c.ImageModel) > 128 || strings.ContainsAny(c.ImageModel, "\r\n") || len(c.OpenAIModel) > 128 || len(c.LiveModel) > 128 || len(c.LiveVoice) > 64 {
		return Config{}, errors.New("invalid private OpenAI configuration")
	}
	var value int
	settings := []struct {
		name               string
		fallback, min, max int
		destination        *int
	}{
		{"YAFA_AGENT_MAX_TURNS", 6, 1, 12, &c.AgentLimits.MaxTurns},
		{"YAFA_AGENT_MAX_TOOL_CALLS", 12, 1, 32, &c.AgentLimits.MaxToolCalls},
		{"YAFA_AGENT_MAX_OUTPUT_TOKENS", 2048, 128, 8192, &c.AgentLimits.MaxOutputTokens},
		{"YAFA_AGENT_MAX_INPUT_BYTES", 48000, 4096, 65536, &c.AgentLimits.MaxInputBytes},
		{"YAFA_AGENT_MAX_RUN_SECONDS", 90, 10, 180, &c.AgentLimits.MaxDurationSeconds},
		{"YAFA_LIVE_MAX_SECONDS", 180, 30, 600, &c.LiveMaxSeconds},
		{"YAFA_LIVE_IDLE_SECONDS", 60, 15, 300, &c.LiveIdleSeconds},
	}
	for _, s := range settings {
		value = s.fallback
		if text := os.Getenv(s.name); text != "" {
			var e error
			value, e = strconv.Atoi(text)
			if e != nil {
				return Config{}, errors.New("invalid " + s.name)
			}
		}
		if value < s.min || value > s.max {
			return Config{}, errors.New("out-of-range " + s.name)
		}
		*s.destination = value
	}
	c.AgentLimits.MaxTokens = 24000
	if text := os.Getenv("YAFA_AGENT_MAX_TOKENS"); text != "" {
		n, e := strconv.ParseInt(text, 10, 64)
		if e != nil || n < 1024 || n > 100000 {
			return Config{}, errors.New("invalid YAFA_AGENT_MAX_TOKENS")
		}
		c.AgentLimits.MaxTokens = n
	}
	if int64(c.AgentLimits.MaxOutputTokens) > c.AgentLimits.MaxTokens {
		return Config{}, errors.New("output token limit exceeds run token threshold")
	}
	prices := []string{"OPENAI_INPUT_USD_PER_MILLION", "OPENAI_CACHED_INPUT_USD_PER_MILLION", "OPENAI_OUTPUT_USD_PER_MILLION"}
	present := 0
	for _, name := range prices {
		if os.Getenv(name) != "" {
			present++
		}
	}
	if present != 0 && present != 3 {
		return Config{}, errors.New("configure all three token prices or leave all empty")
	}
	if present == 3 {
		for _, name := range prices {
			n, e := strconv.ParseFloat(os.Getenv(name), 64)
			if e != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 || n > 1000000 {
				return Config{}, errors.New("invalid token price")
			}
			c.TokenPrices = append(c.TokenPrices, n)
		}
	}
	if len(c.TokenPrices) == 0 && c.OpenAIModel == "gpt-6-luna" {
		c.TokenPrices = []float64{0.10, 0.01, 0.50}
	}
	if c.LiveModel == "gpt-live-1" {
		price := 0.05
		c.LivePrice = &price
	}
	if text := os.Getenv("OPENAI_LIVE_USD_PER_MINUTE"); text != "" {
		price, err := strconv.ParseFloat(text, 64)
		if err != nil || price < 0 || price > 1000 || math.IsNaN(price) || math.IsInf(price, 0) {
			return Config{}, errors.New("invalid live price")
		}
		c.LivePrice = &price
	}
	for _, flag := range []struct {
		name   string
		target *bool
	}{
		{"YAFA_CODEX_ENABLED", &c.Features.Codex}, {"YAFA_SPOTIFY_ENABLED", &c.Features.Spotify},
		{"YAFA_IMAGES_ENABLED", &c.Features.Images}, {"YAFA_ANALYSIS_ENABLED", &c.Features.Analysis}, {"YAFA_LIVE_ENABLED", &c.Features.Live},
	} {
		value := os.Getenv(flag.name)
		if value != "" && value != "true" && value != "false" {
			return Config{}, errors.New(flag.name + " must be true or false")
		}
		*flag.target = value == "true"
	}
	c.CodexHost, c.CodexUser = os.Getenv("YAFA_CODEX_SSH_HOST"), os.Getenv("YAFA_CODEX_SSH_USER")
	c.CodexKeyFile, c.CodexKnownHostsFile = os.Getenv("YAFA_CODEX_SSH_KEY_FILE"), os.Getenv("YAFA_CODEX_SSH_KNOWN_HOSTS_FILE")
	c.CodexRunner = os.Getenv("YAFA_CODEX_REMOTE_RUNNER")
	c.CodexPort, c.CodexMaxSeconds = 22, 300
	for _, setting := range []struct {
		name     string
		target   *int
		min, max int
	}{
		{"YAFA_CODEX_SSH_PORT", &c.CodexPort, 1, 65535}, {"YAFA_CODEX_MAX_SECONDS", &c.CodexMaxSeconds, 30, 600},
	} {
		if text := os.Getenv(setting.name); text != "" {
			n, e := strconv.Atoi(text)
			if e != nil || n < setting.min || n > setting.max {
				return Config{}, errors.New("invalid " + setting.name)
			}
			*setting.target = n
		}
	}
	enabled := os.Getenv("YAFA_AI_ENABLED")
	if enabled != "" && enabled != "true" && enabled != "false" {
		return Config{}, errors.New("YAFA_AI_ENABLED must be true or false")
	}
	c.Safety = core.SafetyLimits{Enabled: enabled == "true",
		Window5h:  core.SafetyQuota{Runs: 30, Tokens: 100000, VoiceSessions: 10, VoiceSeconds: 900, Hosted: 3, Images: 2, CodexJobs: 2, KnownUSD: 1},
		Window24h: core.SafetyQuota{Runs: 60, Tokens: 250000, VoiceSessions: 20, VoiceSeconds: 1800, Hosted: 6, Images: 3, CodexJobs: 4, KnownUSD: 2}}
	return c, nil
}
