package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	for _, field := range []string{"version:", "commit:", "date:"} {
		if !strings.Contains(stdout.String(), field) {
			t.Fatalf("expected version output to contain %q", field)
		}
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunHelpListsCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	for _, command := range []string{"run", "init", "config-path", "version", "help"} {
		if !strings.Contains(stdout.String(), command) {
			t.Fatalf("expected help output to contain %q", command)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"unknown"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}

	if !strings.Contains(stderr.String(), "unknown command: unknown") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}
