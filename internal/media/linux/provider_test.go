//go:build linux

package linux

import (
	"testing"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

func TestPlaybackStatus(t *testing.T) {
	tests := map[string]media.PlaybackStatus{
		"Playing": media.StatusPlaying,
		"Paused":  media.StatusPaused,
		"Stopped": media.StatusStopped,
		"Other":   media.StatusUnknown,
	}

	for input, expected := range tests {
		if got := playbackStatus(input); got != expected {
			t.Fatalf("expected %q for %q, got %q", expected, input, got)
		}
	}
}
