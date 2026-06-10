//go:build !linux && !windows

package app

import "os"

func terminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
