package windows

import (
	"testing"
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

func TestMapPlaybackStatus(t *testing.T) {
	tests := []struct {
		value int32
		want  media.PlaybackStatus
	}{
		{playbackClosed, media.StatusUnknown},
		{playbackOpened, media.StatusUnknown},
		{playbackChanging, media.StatusUnknown},
		{playbackStopped, media.StatusStopped},
		{playbackPlaying, media.StatusPlaying},
		{playbackPaused, media.StatusPaused},
		{99, media.StatusUnknown},
	}

	for _, test := range tests {
		if got := mapPlaybackStatus(test.value); got != test.want {
			t.Fatalf("mapPlaybackStatus(%d) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestTrackFromUnavailableSnapshot(t *testing.T) {
	track := trackFromSnapshot(snapshot{}, time.Now())
	if track.Status != media.StatusUnknown {
		t.Fatalf("status = %q, want %q", track.Status, media.StatusUnknown)
	}
	if track.Title != "" || track.Artist != "" || track.Position != 0 {
		t.Fatalf("unexpected unavailable track: %#v", track)
	}
}

func TestTrackFromPlayingSnapshotAdvancesPosition(t *testing.T) {
	now := time.Date(2026, time.June, 10, 12, 0, 2, 0, time.UTC)
	updated := now.Add(-2 * time.Second)
	updatedTicks := windowsEpochTicks + updated.UnixNano()/100

	track := trackFromSnapshot(snapshot{
		available:        true,
		title:            "Song",
		artist:           "Artist",
		positionTicks:    int64(30 * time.Second / (100 * time.Nanosecond)),
		lastUpdatedTicks: updatedTicks,
		playbackStatus:   playbackPlaying,
	}, now)

	if track.Position != 32*time.Second {
		t.Fatalf("position = %v, want 32s", track.Position)
	}
	if track.Status != media.StatusPlaying {
		t.Fatalf("status = %q, want %q", track.Status, media.StatusPlaying)
	}
}

func TestTrackFromPausedSnapshotKeepsReportedPosition(t *testing.T) {
	now := time.Date(2026, time.June, 10, 12, 0, 2, 0, time.UTC)
	track := trackFromSnapshot(snapshot{
		available:        true,
		positionTicks:    int64(30 * time.Second / (100 * time.Nanosecond)),
		lastUpdatedTicks: windowsEpochTicks + now.Add(-2*time.Second).UnixNano()/100,
		playbackStatus:   playbackPaused,
	}, now)

	if track.Position != 30*time.Second {
		t.Fatalf("position = %v, want 30s", track.Position)
	}
}
