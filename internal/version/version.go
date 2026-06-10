package version

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("DisGOrd Lyrics\nversion: %s\ncommit: %s\ndate: %s\n", Version, Commit, Date)
}
