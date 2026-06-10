package lyrics

import "regexp"

var timestampPattern = regexp.MustCompile(`\[(\d+):(\d{1,2})(?:\.(\d{1,3}))?\]`)
