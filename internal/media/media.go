package media

import (
	"context"
	"errors"
	"strings"
	"time"
)

type PlaybackStatus string

const (
	StatusPlaying PlaybackStatus = "playing"
	StatusPaused  PlaybackStatus = "paused"
	StatusStopped PlaybackStatus = "stopped"
	StatusUnknown PlaybackStatus = "unknown"
)

var ErrUnsupported = errors.New("media detection is unsupported on this platform")

type Track struct {
	Title    string
	Artist   string
	Position time.Duration
	Status   PlaybackStatus
}

func (track Track) ID() string {
	return strings.ToLower(strings.TrimSpace(track.Title)) + "\x00" + strings.ToLower(strings.TrimSpace(track.Artist))
}

type Provider interface {
	Current(context.Context) (Track, error)
}
