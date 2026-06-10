package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/lyrics"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/status"
)

type statusManager interface {
	Set(context.Context, string) (bool, error)
	Clear(context.Context) (bool, error)
}

type Runtime struct {
	media          media.Provider
	lyrics         lyrics.Provider
	status         statusManager
	output         io.Writer
	logger         *logger
	prefix         string
	maxLength      int
	clearOnPause   bool
	offset         time.Duration
	interval       time.Duration
	lyricsRetry    time.Duration
	requestTimeout time.Duration
	trackID        string
	lines          []lyrics.Line
	lyricsReady    bool
	nextRetry      time.Time
}

func (runtime *Runtime) Run(ctx context.Context) error {
	clearCtx, cancel := context.WithTimeout(ctx, runtime.requestTimeout)
	_, err := runtime.status.Clear(clearCtx)
	cancel()
	if err != nil {
		runtime.logger.warn("clear Discord status on startup: %v", err)
	}

	if err := runtime.poll(ctx, time.Now()); err != nil {
		return err
	}

	ticker := time.NewTicker(runtime.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			if err := runtime.poll(ctx, now); err != nil {
				return err
			}
		}
	}
}

func (runtime *Runtime) poll(ctx context.Context, now time.Time) error {
	mediaCtx, cancel := context.WithTimeout(ctx, runtime.requestTimeout)
	track, err := runtime.media.Current(mediaCtx)
	cancel()
	if err != nil {
		if errors.Is(err, media.ErrUnsupported) {
			return err
		}
		runtime.logger.warn("read current media: %v", err)
		runtime.clearIfConfigured(ctx)
		return nil
	}

	if track.Status != media.StatusPlaying || track.Title == "" {
		runtime.clearIfConfigured(ctx)
		return nil
	}

	trackID := track.ID()
	if trackID != runtime.trackID {
		runtime.trackID = trackID
		runtime.lines = nil
		runtime.lyricsReady = false
		runtime.nextRetry = time.Time{}
		runtime.logger.info("playing %s", displayTrack(track))
	}

	if !runtime.lyricsReady && !now.Before(runtime.nextRetry) {
		fetchCtx, cancel := context.WithTimeout(ctx, runtime.requestTimeout)
		lines, fetchErr := runtime.lyrics.Fetch(fetchCtx, track.Title, track.Artist)
		cancel()

		switch {
		case fetchErr == nil:
			runtime.lines = lines
			runtime.lyricsReady = true
		case errors.Is(fetchErr, lyrics.ErrNotFound):
			runtime.lines = nil
			runtime.lyricsReady = true
			runtime.logger.info("no synchronized lyrics for %s", displayTrack(track))
		default:
			runtime.nextRetry = now.Add(runtime.lyricsRetry)
			runtime.logger.warn("fetch synchronized lyrics: %v", fetchErr)
		}
	}

	line, _ := lyrics.ActiveLine(runtime.lines, track.Position+runtime.offset)
	formatted := status.Format(line, runtime.prefix, runtime.maxLength)

	updateCtx, cancel := context.WithTimeout(ctx, runtime.requestTimeout)
	changed, updateErr := runtime.status.Set(updateCtx, formatted)
	cancel()
	if updateErr != nil {
		runtime.logger.warn("update Discord status: %v", updateErr)
		return nil
	}

	if changed {
		render(runtime.output, track, formatted)
	}

	return nil
}

func (runtime *Runtime) clearIfConfigured(ctx context.Context) {
	if !runtime.clearOnPause {
		return
	}

	clearCtx, cancel := context.WithTimeout(ctx, runtime.requestTimeout)
	_, err := runtime.status.Clear(clearCtx)
	cancel()
	if err != nil {
		runtime.logger.warn("clear Discord status: %v", err)
	}
}

func render(output io.Writer, track media.Track, line string) {
	minutes := int(track.Position / time.Minute)
	seconds := int(track.Position/time.Second) % 60
	if line == "" {
		line = "..."
	}

	fmt.Fprintf(output, "Song: %s\nArtist: %s\nTime: %02d:%02d\nLyrics: %s\n\n", track.Title, track.Artist, minutes, seconds, line)
}
