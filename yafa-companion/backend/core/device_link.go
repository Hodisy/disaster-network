package core

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"time"
	"yafa/backend/entities"
	"yafa/backend/repositories"
)

var DeviceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,31}$`)
var bootIDPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)

type DeviceLink struct {
	Repository repositories.DeviceLinkRepository
	DeviceID   string
	Now        func() time.Time
}

func validExpression(value string) bool {
	switch value {
	case "idle", "waking", "listening", "thinking", "speaking", "happy", "wink", "sad", "reminder", "sleepy":
		return true
	}
	return false
}

func validAudio(value string) bool {
	switch value {
	case "unavailable", "idle", "tone", "recording", "ready", "playing", "failed", "loading", "exporting":
		return true
	}
	return false
}

func (c DeviceLink) Read(ctx context.Context, id string) (entities.DeviceLink, bool, error) {
	if c.Repository == nil || c.Now == nil {
		return entities.DeviceLink{}, false, ErrNotConfigured
	}
	if id != c.DeviceID || !DeviceIDPattern.MatchString(id) {
		return entities.DeviceLink{}, false, ErrDeviceAccessDenied
	}
	link, _, err := c.Repository.ReadLink(ctx, id)
	age := c.Now().Sub(link.LastSeenAt)
	return link, err == nil && !link.LastSeenAt.IsZero() && age >= 0 && age <= 45*time.Second, err
}

func (c DeviceLink) Exchange(ctx context.Context, report entities.DeviceReport) (entities.DeviceExchange, error) {
	if c.Repository == nil || c.Now == nil {
		return entities.DeviceExchange{}, ErrNotConfigured
	}
	if report.DeviceID != c.DeviceID || !DeviceIDPattern.MatchString(report.DeviceID) {
		return entities.DeviceExchange{}, ErrDeviceAccessDenied
	}
	applied, err := strconv.ParseUint(report.AppliedVersion, 10, 64)
	if err != nil || !bootIDPattern.MatchString(report.BootID) || len(report.Firmware) == 0 || len(report.Firmware) > 48 || !report.Display || !validExpression(report.Expression) || !validAudio(report.Audio) {
		return entities.DeviceExchange{}, ErrInvalidRecord
	}
	for attempt := 0; attempt < 3; attempt++ {
		link, version, err := c.Repository.ReadLink(ctx, c.DeviceID)
		newLink := errors.Is(err, ErrRecordNotFound)
		if newLink {
			link = entities.DeviceLink{DeviceID: c.DeviceID, Desired: entities.DeviceDisplayCommand{Version: "1", Expression: "idle", Awake: true}}
			version = "0"
		} else if err != nil {
			return entities.DeviceExchange{}, err
		}
		desired, err := strconv.ParseUint(link.Desired.Version, 10, 64)
		if err != nil || applied > desired {
			return entities.DeviceExchange{}, ErrInvalidRecord
		}
		now := c.Now().UTC()
		// Persist changes immediately and unchanged heartbeats at most every 15s.
		if newLink || now.Sub(link.LastSeenAt) >= 15*time.Second || report.BootID != link.Report.BootID || report.AppliedVersion != link.Report.AppliedVersion || report.Expression != link.Report.Expression || report.Awake != link.Report.Awake || report.Audio != link.Report.Audio {
			link.LastSeenAt = now
			link.Report = report
			if err := c.Repository.WriteLink(ctx, c.DeviceID, version, link); err != nil {
				if errors.Is(err, ErrVersionConflict) {
					continue
				}
				return entities.DeviceExchange{}, err
			}
		}
		return entities.DeviceExchange{Protocol: 1, DeviceID: c.DeviceID, Desired: link.Desired}, nil
	}
	return entities.DeviceExchange{}, ErrVersionConflict
}

func (c DeviceLink) SetPresentation(ctx context.Context, id, expected, expression string, awake bool) (entities.DeviceDisplayCommand, error) {
	if c.Repository == nil || c.Now == nil {
		return entities.DeviceDisplayCommand{}, ErrNotConfigured
	}
	if id != c.DeviceID || !DeviceIDPattern.MatchString(id) {
		return entities.DeviceDisplayCommand{}, ErrDeviceAccessDenied
	}
	if !validExpression(expression) {
		return entities.DeviceDisplayCommand{}, ErrInvalidRecord
	}
	for attempt := 0; attempt < 3; attempt++ {
		link, version, err := c.Repository.ReadLink(ctx, id)
		if err != nil {
			return entities.DeviceDisplayCommand{}, err
		}
		if expected != link.Desired.Version {
			return entities.DeviceDisplayCommand{}, ErrVersionConflict
		}
		n, err := strconv.ParseUint(expected, 10, 64)
		if err != nil || n == 0 || n == math.MaxUint64 {
			return entities.DeviceDisplayCommand{}, ErrInvalidRecord
		}
		link.Desired = entities.DeviceDisplayCommand{Version: strconv.FormatUint(n+1, 10), Expression: expression, Awake: awake}
		if err := c.Repository.WriteLink(ctx, id, version, link); err != nil {
			if errors.Is(err, ErrVersionConflict) {
				continue
			}
			return entities.DeviceDisplayCommand{}, err
		}
		return link.Desired, nil
	}
	return entities.DeviceDisplayCommand{}, ErrVersionConflict
}
