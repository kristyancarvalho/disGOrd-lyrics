package windows

import (
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

const (
	playbackClosed    int32 = 0
	playbackOpened    int32 = 1
	playbackChanging  int32 = 2
	playbackStopped   int32 = 3
	playbackPlaying   int32 = 4
	playbackPaused    int32 = 5
	windowsEpochTicks       = int64(116444736000000000)
)

type snapshot struct {
	available        bool
	title            string
	artist           string
	positionTicks    int64
	lastUpdatedTicks int64
	playbackStatus   int32
}

func trackFromSnapshot(value snapshot, now time.Time) media.Track {
	if !value.available {
		return media.Track{Status: media.StatusUnknown}
	}

	status := mapPlaybackStatus(value.playbackStatus)
	position := ticksToDuration(value.positionTicks)
	if status == media.StatusPlaying {
		position += elapsedSince(value.lastUpdatedTicks, now)
	}
	if position < 0 {
		position = 0
	}

	return media.Track{
		Title:    value.title,
		Artist:   value.artist,
		Position: position,
		Status:   status,
	}
}

func mapPlaybackStatus(value int32) media.PlaybackStatus {
	switch value {
	case playbackPlaying:
		return media.StatusPlaying
	case playbackPaused:
		return media.StatusPaused
	case playbackStopped:
		return media.StatusStopped
	default:
		return media.StatusUnknown
	}
}

func ticksToDuration(value int64) time.Duration {
	return time.Duration(value) * 100 * time.Nanosecond
}

func elapsedSince(value int64, now time.Time) time.Duration {
	if value <= windowsEpochTicks {
		return 0
	}

	updated := time.Unix(0, (value-windowsEpochTicks)*100).UTC()
	elapsed := now.UTC().Sub(updated)
	if elapsed < 0 || elapsed > 24*time.Hour {
		return 0
	}
	return elapsed
}
