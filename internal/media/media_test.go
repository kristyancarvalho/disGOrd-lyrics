package media

import (
	"testing"
	"time"
)

func TestTrackID(t *testing.T) {
	first := Track{Title: " Song ", Artist: "ARTIST", Position: time.Second}
	second := Track{Title: "song", Artist: "artist", Position: time.Hour}

	if first.ID() != second.ID() {
		t.Fatalf("expected stable track IDs: %q %q", first.ID(), second.ID())
	}
}
