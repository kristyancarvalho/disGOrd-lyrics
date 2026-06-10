//go:build !linux && !windows

package app

import (
	"fmt"
	"runtime"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

func newMediaProvider() (media.Provider, error) {
	return nil, fmt.Errorf("%w: %s", media.ErrUnsupported, runtime.GOOS)
}
