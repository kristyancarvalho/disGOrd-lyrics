package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"strings"
	"time"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/config"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/discord"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/lyrics"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/status"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/version"
)

const usage = `DisGOrd Lyrics

Usage:
  disgord-lyrics run
  disgord-lyrics init [--force]
  disgord-lyrics config-path
  disgord-lyrics version
  disgord-lyrics help

Examples:
  disgord-lyrics init
  disgord-lyrics config-path
  disgord-lyrics run
`

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return 0
	}

	switch args[0] {
	case "run":
		if len(args) != 1 {
			return commandError(stderr, "run does not accept arguments")
		}
		return runCommand(stdout, stderr)
	case "init":
		return initCommand(args[1:], stdout, stderr)
	case "config-path":
		if len(args) != 1 {
			return commandError(stderr, "config-path does not accept arguments")
		}
		return configPathCommand(stdout, stderr)
	case "version":
		if len(args) != 1 {
			return commandError(stderr, "version does not accept arguments")
		}
		fmt.Fprint(stdout, version.String())
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		return commandError(stderr, fmt.Sprintf("unknown command: %s", args[0]))
	}
}

func initCommand(args []string, stdout, stderr io.Writer) int {
	force := false
	for _, arg := range args {
		if arg != "--force" {
			return commandError(stderr, fmt.Sprintf("unknown init option: %s", arg))
		}
		force = true
	}

	path, err := config.Path()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return 1
	}
	if err := config.Init(path, force); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintln(stdout, path)
	return 0
}

func configPathCommand(stdout, stderr io.Writer) int {
	path, err := config.Path()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, path)
	return 0
}

func runCommand(stdout, stderr io.Writer) int {
	path, err := config.Path()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return 1
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	mediaProvider, err := newMediaProvider()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	lyricsProvider := lyrics.NewCached(lyrics.NewLRCLIB(&http.Client{Timeout: 10 * time.Second}, version.Version))
	statusManager := status.NewManager(discord.New(cfg.Discord.Token))
	logger := newLogger(stderr, cfg.Logging.Level)

	runtime := &Runtime{
		media:          mediaProvider,
		lyrics:         lyricsProvider,
		status:         statusManager,
		output:         stdout,
		logger:         logger,
		prefix:         cfg.Status.Prefix,
		maxLength:      cfg.Status.MaxLength,
		clearOnPause:   cfg.Status.ClearOnPause,
		offset:         time.Duration(cfg.Lyrics.OffsetMS) * time.Millisecond,
		interval:       time.Duration(cfg.Polling.IntervalMS) * time.Millisecond,
		lyricsRetry:    15 * time.Second,
		requestTimeout: 6 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), terminationSignals()...)
	defer stop()

	if cfg.Status.ClearOnExit {
		defer func() {
			clearCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			if _, err := statusManager.Clear(clearCtx); err != nil {
				logger.warn("clear Discord status on exit: %v", err)
			}
		}()
	}

	if err := runtime.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func commandError(stderr io.Writer, message string) int {
	fmt.Fprintln(stderr, message)
	fmt.Fprint(stderr, usage)
	return 2
}

type logger struct {
	writer io.Writer
	level  int
}

func newLogger(writer io.Writer, level string) *logger {
	levels := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	return &logger{writer: writer, level: levels[level]}
}

func (logger *logger) info(format string, args ...any) {
	logger.write(1, "info", format, args...)
}

func (logger *logger) warn(format string, args ...any) {
	logger.write(2, "warn", format, args...)
}

func (logger *logger) write(level int, name, format string, args ...any) {
	if level < logger.level {
		return
	}
	fmt.Fprintf(logger.writer, name+": "+format+"\n", args...)
}

func displayTrack(track media.Track) string {
	title := strings.TrimSpace(track.Title)
	artist := strings.TrimSpace(track.Artist)
	if artist == "" {
		return title
	}
	return title + " - " + artist
}
