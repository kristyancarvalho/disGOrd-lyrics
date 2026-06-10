package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultLRCLIBURL = "https://lrclib.net/api/search"

var ErrNotFound = errors.New("synchronized lyrics not found")

type Provider interface {
	Fetch(context.Context, string, string) ([]Line, error)
}

type LRCLIB struct {
	client    *http.Client
	endpoint  string
	userAgent string
}

type lrclibRecord struct {
	TrackName    string `json:"trackName"`
	ArtistName   string `json:"artistName"`
	Instrumental bool   `json:"instrumental"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func NewLRCLIB(client *http.Client, version string) *LRCLIB {
	return NewLRCLIBWithURL(client, version, defaultLRCLIBURL)
}

func NewLRCLIBWithURL(client *http.Client, version, endpoint string) *LRCLIB {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if version == "" {
		version = "dev"
	}
	return &LRCLIB{
		client:    client,
		endpoint:  endpoint,
		userAgent: "disgord-lyrics/" + version + " (https://github.com/kristyancarvalho/disGOrd-lyrics)",
	}
}

func (provider *LRCLIB) Fetch(ctx context.Context, title, artist string) ([]Line, error) {
	requestURL, err := url.Parse(provider.endpoint)
	if err != nil {
		return nil, fmt.Errorf("prepare LRCLIB request: %w", err)
	}

	query := requestURL.Query()
	query.Set("track_name", title)
	query.Set("artist_name", artist)
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("prepare LRCLIB request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", provider.userAgent)

	response, err := provider.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request LRCLIB lyrics: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("request LRCLIB lyrics: unexpected HTTP status %d", response.StatusCode)
	}

	var records []lrclibRecord
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&records); err != nil {
		return nil, fmt.Errorf("decode LRCLIB response: %w", err)
	}

	record, ok := selectRecord(records, title, artist)
	if !ok {
		return nil, ErrNotFound
	}

	lines := ParseLRC(record.SyncedLyrics)
	if len(lines) == 0 {
		return nil, ErrNotFound
	}

	return lines, nil
}

func selectRecord(records []lrclibRecord, title, artist string) (lrclibRecord, bool) {
	for _, record := range records {
		if record.Instrumental || strings.TrimSpace(record.SyncedLyrics) == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(record.TrackName), strings.TrimSpace(title)) &&
			strings.EqualFold(strings.TrimSpace(record.ArtistName), strings.TrimSpace(artist)) {
			return record, true
		}
	}

	for _, record := range records {
		if !record.Instrumental && strings.TrimSpace(record.SyncedLyrics) != "" {
			return record, true
		}
	}

	return lrclibRecord{}, false
}

type Cached struct {
	provider Provider
	mu       sync.Mutex
	entries  map[string]cacheEntry
}

type cacheEntry struct {
	lines []Line
	err   error
}

func NewCached(provider Provider) *Cached {
	return &Cached{
		provider: provider,
		entries:  make(map[string]cacheEntry),
	}
}

func (cache *Cached) Fetch(ctx context.Context, title, artist string) ([]Line, error) {
	key := strings.ToLower(strings.TrimSpace(title)) + "\x00" + strings.ToLower(strings.TrimSpace(artist))

	cache.mu.Lock()
	entry, ok := cache.entries[key]
	cache.mu.Unlock()
	if ok {
		return cloneLines(entry.lines), entry.err
	}

	lines, err := cache.provider.Fetch(ctx, title, artist)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	cache.mu.Lock()
	cache.entries[key] = cacheEntry{lines: cloneLines(lines), err: err}
	cache.mu.Unlock()

	return cloneLines(lines), err
}

func cloneLines(lines []Line) []Line {
	return append([]Line(nil), lines...)
}
