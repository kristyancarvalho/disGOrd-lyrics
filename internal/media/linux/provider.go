//go:build linux

package linux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

const playerInterface = "org.mpris.MediaPlayer2.Player"

var playerPath = dbus.ObjectPath("/org/mpris/MediaPlayer2")

type Provider struct {
	connection *dbus.Conn
}

func New() (*Provider, error) {
	connection, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("connect to D-Bus session: %w", err)
	}
	return &Provider{connection: connection}, nil
}

func (provider *Provider) Current(ctx context.Context) (media.Track, error) {
	var names []string
	call := provider.connection.BusObject().CallWithContext(ctx, "org.freedesktop.DBus.ListNames", 0)
	if err := call.Store(&names); err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("list MPRIS players: %w", err)
	}

	var fallback *media.Track
	for _, name := range names {
		if !strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			continue
		}

		track, err := provider.readTrack(ctx, name)
		if err != nil {
			continue
		}
		if track.Status == media.StatusPlaying {
			return track, nil
		}
		if fallback == nil || track.Status == media.StatusPaused {
			copy := track
			fallback = &copy
		}
	}

	if fallback != nil {
		return *fallback, nil
	}
	return media.Track{Status: media.StatusUnknown}, nil
}

func (provider *Provider) readTrack(ctx context.Context, name string) (media.Track, error) {
	object := provider.connection.Object(name, playerPath)

	statusVariant, err := property(ctx, object, playerInterface+".PlaybackStatus")
	if err != nil {
		return media.Track{}, err
	}

	statusText, ok := statusVariant.Value().(string)
	if !ok {
		return media.Track{}, fmt.Errorf("invalid MPRIS playback status")
	}

	track := media.Track{Status: playbackStatus(statusText)}

	metadataVariant, err := property(ctx, object, playerInterface+".Metadata")
	if err != nil {
		return media.Track{}, err
	}
	metadata, ok := metadataVariant.Value().(map[string]dbus.Variant)
	if !ok {
		return media.Track{}, fmt.Errorf("invalid MPRIS metadata")
	}

	if title, ok := metadata["xesam:title"]; ok {
		track.Title, _ = title.Value().(string)
	}
	if artist, ok := metadata["xesam:artist"]; ok {
		switch value := artist.Value().(type) {
		case []string:
			track.Artist = strings.Join(value, ", ")
		case string:
			track.Artist = value
		}
	}

	positionVariant, err := property(ctx, object, playerInterface+".Position")
	if err == nil {
		switch position := positionVariant.Value().(type) {
		case int64:
			track.Position = time.Duration(position) * time.Microsecond
		case uint64:
			track.Position = time.Duration(position) * time.Microsecond
		}
	}

	return track, nil
}

func property(ctx context.Context, object dbus.BusObject, name string) (dbus.Variant, error) {
	var variant dbus.Variant
	call := object.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, playerInterface, strings.TrimPrefix(name, playerInterface+"."))
	if err := call.Store(&variant); err != nil {
		return dbus.Variant{}, err
	}
	return variant, nil
}

func playbackStatus(status string) media.PlaybackStatus {
	switch strings.ToLower(status) {
	case "playing":
		return media.StatusPlaying
	case "paused":
		return media.StatusPaused
	case "stopped":
		return media.StatusStopped
	default:
		return media.StatusUnknown
	}
}
