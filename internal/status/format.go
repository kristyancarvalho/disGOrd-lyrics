package status

import (
	"regexp"
	"strings"
	"unicode"
)

const maxInputLength = 80

var malformedJoinedWords = regexp.MustCompile(`[a-z][A-Z][a-z]`)
var joinedWords = regexp.MustCompile(`([a-z])([A-Z])`)

func Clean(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\r", ""))
	if text == "" {
		return ""
	}
	if strings.Contains(text, "/") && len([]rune(text)) > 40 {
		return ""
	}
	if malformedJoinedWords.MatchString(text) {
		return ""
	}
	if len([]rune(text)) > maxInputLength {
		return ""
	}

	total := 0
	invalid := 0
	for _, r := range text {
		total++
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			invalid++
		}
	}
	if total > 0 && float64(invalid) > float64(total)*0.4 {
		return ""
	}

	return text
}

func FixJoinedWords(text string) string {
	return joinedWords.ReplaceAllString(text, "$1 $2")
}

func Format(text, prefix string, maxLength int) string {
	text = Clean(text)
	if text == "" || maxLength < 1 {
		return ""
	}

	return truncate(prefix+FixJoinedWords(text), maxLength)
}

func truncate(text string, maxLength int) string {
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength])
}
