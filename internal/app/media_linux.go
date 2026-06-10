//go:build linux

package app

import (
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	linuxmedia "github.com/kristyancarvalho/disGOrd-lyrics/internal/media/linux"
)

func newMediaProvider() (media.Provider, error) {
	return linuxmedia.New()
}
