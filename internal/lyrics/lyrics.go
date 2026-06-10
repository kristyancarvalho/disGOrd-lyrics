package lyrics

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	At   time.Duration
	Text string
}

func ParseLRC(input string) []Line {
	var lines []Line

	for _, raw := range strings.Split(input, "\n") {
		matches := timestampPattern.FindAllStringSubmatchIndex(raw, -1)
		if len(matches) == 0 {
			continue
		}

		textStart := matches[len(matches)-1][1]
		text := strings.TrimSpace(raw[textStart:])

		for _, match := range matches {
			minutes, err := strconv.Atoi(raw[match[2]:match[3]])
			if err != nil {
				continue
			}
			seconds, err := strconv.Atoi(raw[match[4]:match[5]])
			if err != nil || seconds > 59 {
				continue
			}

			milliseconds := 0
			if match[6] >= 0 {
				fraction := raw[match[6]:match[7]]
				switch len(fraction) {
				case 1:
					milliseconds, _ = strconv.Atoi(fraction + "00")
				case 2:
					milliseconds, _ = strconv.Atoi(fraction + "0")
				case 3:
					milliseconds, _ = strconv.Atoi(fraction)
				}
			}

			at := time.Duration(minutes)*time.Minute +
				time.Duration(seconds)*time.Second +
				time.Duration(milliseconds)*time.Millisecond
			lines = append(lines, Line{At: at, Text: text})
		}
	}

	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].At < lines[j].At
	})

	return lines
}

func ActiveLine(lines []Line, position time.Duration) (string, bool) {
	if len(lines) == 0 {
		return "", false
	}

	index := sort.Search(len(lines), func(i int) bool {
		return lines[i].At > position
	})
	if index == 0 {
		return "", false
	}

	return lines[index-1].Text, true
}
