package app

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/lyrics"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/status"
)

type fakeMedia struct {
	track media.Track
	err   error
}

func (provider *fakeMedia) Current(context.Context) (media.Track, error) {
	return provider.track, provider.err
}

type fakeLyrics struct {
	lines []lyrics.Line
	err   error
	calls int
}

func (provider *fakeLyrics) Fetch(context.Context, string, string) ([]lyrics.Line, error) {
	provider.calls++
	return provider.lines, provider.err
}

type fakeDiscord struct {
	updates []string
	clears  int
}

func (client *fakeDiscord) Update(_ context.Context, text string) error {
	client.updates = append(client.updates, text)
	return nil
}

func (client *fakeDiscord) Clear(context.Context) error {
	client.clears++
	return nil
}

func TestRuntimePollUpdatesActiveLyricOnce(t *testing.T) {
	mediaProvider := &fakeMedia{track: media.Track{
		Title:    "Song",
		Artist:   "Artist",
		Position: 2 * time.Second,
		Status:   media.StatusPlaying,
	}}
	lyricsProvider := &fakeLyrics{lines: []lyrics.Line{
		{At: time.Second, Text: "First line"},
		{At: 3 * time.Second, Text: "Second line"},
	}}
	discordClient := &fakeDiscord{}
	manager := status.NewManager(discordClient)
	var output bytes.Buffer
	var logs bytes.Buffer

	runtime := &Runtime{
		media:          mediaProvider,
		lyrics:         lyricsProvider,
		status:         manager,
		output:         &output,
		logger:         newLogger(&logs, "debug"),
		prefix:         "Now: ",
		maxLength:      70,
		clearOnPause:   true,
		offset:         500 * time.Millisecond,
		lyricsRetry:    15 * time.Second,
		requestTimeout: time.Second,
	}

	now := time.Now()
	if err := runtime.poll(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := runtime.poll(context.Background(), now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if lyricsProvider.calls != 1 {
		t.Fatalf("expected one lyrics request, got %d", lyricsProvider.calls)
	}
	if len(discordClient.updates) != 1 || discordClient.updates[0] != "Now: First line" {
		t.Fatalf("unexpected Discord updates: %#v", discordClient.updates)
	}
	if output.Len() == 0 {
		t.Fatal("expected terminal output")
	}
}

func TestRuntimePollClearsPausedStatus(t *testing.T) {
	mediaProvider := &fakeMedia{track: media.Track{Status: media.StatusPaused}}
	discordClient := &fakeDiscord{}
	manager := status.NewManager(discordClient)

	runtime := &Runtime{
		media:          mediaProvider,
		lyrics:         &fakeLyrics{},
		status:         manager,
		output:         &bytes.Buffer{},
		logger:         newLogger(&bytes.Buffer{}, "debug"),
		clearOnPause:   true,
		requestTimeout: time.Second,
	}

	if err := runtime.poll(context.Background(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.poll(context.Background(), time.Now()); err != nil {
		t.Fatal(err)
	}

	if discordClient.clears != 1 {
		t.Fatalf("expected one clear request, got %d", discordClient.clears)
	}
}

func TestRuntimePollReturnsUnsupportedMediaError(t *testing.T) {
	runtime := &Runtime{
		media:          &fakeMedia{err: media.ErrUnsupported},
		lyrics:         &fakeLyrics{},
		status:         status.NewManager(&fakeDiscord{}),
		output:         &bytes.Buffer{},
		logger:         newLogger(&bytes.Buffer{}, "debug"),
		requestTimeout: time.Second,
	}

	err := runtime.poll(context.Background(), time.Now())
	if !errors.Is(err, media.ErrUnsupported) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}
