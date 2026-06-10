//go:build windows

package app

import (
	"fmt"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

func newMediaProvider() (media.Provider, error) {
	return nil, fmt.Errorf("%w: Windows system media sessions are not available in this release", media.ErrUnsupported)
}
