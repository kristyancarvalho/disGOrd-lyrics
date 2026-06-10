package lyrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseLRCSortsAndParsesTimestamps(t *testing.T) {
	input := "[01:02.345] third\n[00:03] first\n[00:04.25] second\n[bad] ignored"
	lines := ParseLRC(input)

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0].At != 3*time.Second || lines[0].Text != "first" {
		t.Fatalf("unexpected first line: %#v", lines[0])
	}
	if lines[1].At != 4250*time.Millisecond || lines[1].Text != "second" {
		t.Fatalf("unexpected second line: %#v", lines[1])
	}
	if lines[2].At != 62*time.Second+345*time.Millisecond || lines[2].Text != "third" {
		t.Fatalf("unexpected third line: %#v", lines[2])
	}
}

func TestParseLRCMultipleTimestampsAndEmptyInput(t *testing.T) {
	lines := ParseLRC("[00:01.00][00:02.000] repeated")
	if len(lines) != 2 || lines[0].Text != "repeated" || lines[1].Text != "repeated" {
		t.Fatalf("unexpected repeated lines: %#v", lines)
	}
	if lines := ParseLRC(""); len(lines) != 0 {
		t.Fatalf("expected no lines, got %#v", lines)
	}
}

func TestActiveLine(t *testing.T) {
	lines := []Line{
		{At: time.Second, Text: "one"},
		{At: 2 * time.Second, Text: "two"},
	}

	if _, ok := ActiveLine(lines, 500*time.Millisecond); ok {
		t.Fatal("expected no active line")
	}
	if line, ok := ActiveLine(lines, 2500*time.Millisecond); !ok || line != "two" {
		t.Fatalf("expected second line, got %q %v", line, ok)
	}
}

func TestLRCLIBFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("track_name") != "Song" || request.URL.Query().Get("artist_name") != "Artist" {
			t.Fatalf("unexpected query: %s", request.URL.RawQuery)
		}
		if !strings.HasPrefix(request.Header.Get("User-Agent"), "disgord-lyrics/") {
			t.Fatalf("unexpected user agent: %q", request.Header.Get("User-Agent"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[{"trackName":"Song","artistName":"Artist","instrumental":false,"syncedLyrics":"[00:01.00] line"}]`))
	}))
	defer server.Close()

	provider := NewLRCLIBWithURL(server.Client(), "test", server.URL)
	lines, err := provider.Fetch(context.Background(), "Song", "Artist")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0].Text != "line" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestLRCLIBNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`[]`))
	}))
	defer server.Close()

	provider := NewLRCLIBWithURL(server.Client(), "test", server.URL)
	_, err := provider.Fetch(context.Background(), "Song", "Artist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

type countingProvider struct {
	calls int
}

func (provider *countingProvider) Fetch(context.Context, string, string) ([]Line, error) {
	provider.calls++
	return []Line{{At: time.Second, Text: "line"}}, nil
}

func TestCachedProvider(t *testing.T) {
	source := &countingProvider{}
	cache := NewCached(source)

	for range 2 {
		if _, err := cache.Fetch(context.Background(), "Song", "Artist"); err != nil {
			t.Fatal(err)
		}
	}

	if source.calls != 1 {
		t.Fatalf("expected one source call, got %d", source.calls)
	}
}
