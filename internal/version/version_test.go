package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	output := String()

	for _, field := range []string{"DisGOrd Lyrics", "version:", "commit:", "date:"} {
		if !strings.Contains(output, field) {
			t.Fatalf("expected output to contain %q", field)
		}
	}
}
