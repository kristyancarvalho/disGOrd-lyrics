//go:build windows

package app

import (
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	windowsmedia "github.com/kristyancarvalho/disGOrd-lyrics/internal/media/windows"
)

func newMediaProvider() (media.Provider, error) {
	return windowsmedia.New()
}
