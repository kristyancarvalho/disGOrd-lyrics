//go:build windows

package windows

import (
	"context"
	"testing"
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

func TestProviderCurrentOnWindows(t *testing.T) {
	provider, err := New()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	track, err := provider.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}

	switch track.Status {
	case media.StatusPlaying, media.StatusPaused, media.StatusStopped, media.StatusUnknown:
	default:
		t.Fatalf("unexpected playback status %q", track.Status)
	}
}
