package app

import (
	"fmt"
	"io"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/version"
)

const usage = `DisGOrd Lyrics

Usage:
  disgord-lyrics version
  disgord-lyrics help
`

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return 0
	}

	switch args[0] {
	case "version":
		fmt.Fprint(stdout, version.String())
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		fmt.Fprint(stderr, usage)
		return 2
	}
}
