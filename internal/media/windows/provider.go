//go:build windows

package windows

import (
	"context"
	"fmt"

	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (provider *Provider) Current(context.Context) (media.Track, error) {
	return media.Track{Status: media.StatusUnknown}, fmt.Errorf("%w: Windows system media sessions are not available in this release", media.ErrUnsupported)
}
