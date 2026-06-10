package status

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: " \r\n", want: ""},
		{name: "trim", input: "\r Hello world \r", want: "Hello world"},
		{name: "long slash", input: strings.Repeat("a", 41) + "/", want: ""},
		{name: "malformed joined", input: "helloWorld", want: ""},
		{name: "too long", input: strings.Repeat("a", 81), want: ""},
		{name: "too many symbols", input: "!!!! hello !!!!", want: ""},
		{name: "unicode letters", input: "Olá mundo", want: "Olá mundo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Clean(test.input); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestFixJoinedWords(t *testing.T) {
	if got := FixJoinedWords("helloWorld Again"); got != "hello World Again" {
		t.Fatalf("unexpected fixed text: %q", got)
	}
}

func TestFormat(t *testing.T) {
	if got := Format("Hello world", "🎵 ", 7); got != "🎵 Hello" {
		t.Fatalf("unexpected formatted status: %q", got)
	}
	if got := Format("", "🎵 ", 70); got != "" {
		t.Fatalf("expected empty status, got %q", got)
	}
}

type recordingClient struct {
	updates []string
	clears  int
	err     error
}

func (client *recordingClient) Update(_ context.Context, text string) error {
	client.updates = append(client.updates, text)
	return client.err
}

func (client *recordingClient) Clear(context.Context) error {
	client.clears++
	return client.err
}

func TestManagerSuppressesDuplicates(t *testing.T) {
	client := &recordingClient{}
	manager := NewManager(client)

	changed, err := manager.Set(context.Background(), "hello")
	if err != nil || !changed {
		t.Fatalf("expected first update, changed=%v err=%v", changed, err)
	}
	changed, err = manager.Set(context.Background(), "hello")
	if err != nil || changed {
		t.Fatalf("expected duplicate suppression, changed=%v err=%v", changed, err)
	}
	changed, err = manager.Clear(context.Background())
	if err != nil || !changed {
		t.Fatalf("expected clear, changed=%v err=%v", changed, err)
	}
	changed, err = manager.Clear(context.Background())
	if err != nil || changed {
		t.Fatalf("expected duplicate clear suppression, changed=%v err=%v", changed, err)
	}

	if len(client.updates) != 1 || client.clears != 1 {
		t.Fatalf("unexpected calls: updates=%#v clears=%d", client.updates, client.clears)
	}
}

func TestManagerBacksOffAfterFailure(t *testing.T) {
	client := &recordingClient{err: errors.New("request failed")}
	manager := NewManager(client)
	now := time.Now()
	manager.now = func() time.Time { return now }

	if _, err := manager.Set(context.Background(), "hello"); err == nil {
		t.Fatal("expected first request error")
	}
	if changed, err := manager.Set(context.Background(), "hello"); err != nil || changed {
		t.Fatalf("expected suppressed retry, changed=%v err=%v", changed, err)
	}
	if len(client.updates) != 1 {
		t.Fatalf("expected one request during backoff, got %d", len(client.updates))
	}

	now = now.Add(manager.retryDelay)
	if _, err := manager.Set(context.Background(), "hello"); err == nil {
		t.Fatal("expected retry error")
	}
	if len(client.updates) != 2 {
		t.Fatalf("expected request after backoff, got %d", len(client.updates))
	}
}
