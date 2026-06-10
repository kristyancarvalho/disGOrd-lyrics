//go:build windows

package windows

import (
	"context"
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/kristyancarvalho/disGOrd-lyrics/internal/media"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media/control"
)

type request struct {
	context  context.Context
	response chan result
}

type result struct {
	track media.Track
	err   error
}

type Provider struct {
	requests chan request
}

func New() (*Provider, error) {
	provider := &Provider{requests: make(chan request)}
	ready := make(chan error, 1)
	go provider.serve(ready)
	if err := <-ready; err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *Provider) Current(ctx context.Context) (media.Track, error) {
	response := make(chan result, 1)
	call := request{context: ctx, response: response}

	select {
	case provider.requests <- call:
	case <-ctx.Done():
		return media.Track{Status: media.StatusUnknown}, ctx.Err()
	}

	select {
	case value := <-response:
		return value.track, value.err
	case <-ctx.Done():
		return media.Track{Status: media.StatusUnknown}, ctx.Err()
	}
}

func (provider *Provider) serve(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		ready <- fmt.Errorf("initialize Windows media controls: %w", err)
		return
	}
	defer ole.CoUninitialize()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	manager, err := requestManager(ctx)
	cancel()
	if err != nil {
		ready <- err
		return
	}
	defer manager.Release()

	ready <- nil
	for call := range provider.requests {
		track, currentErr := safeCurrent(call.context, manager)
		call.response <- result{track: track, err: currentErr}
	}
}

func requestManager(ctx context.Context) (*control.GlobalSystemMediaTransportControlsSessionManager, error) {
	operation, err := control.GlobalSystemMediaTransportControlsSessionManagerRequestAsync()
	if err != nil {
		return nil, fmt.Errorf("request Windows media session manager: %w", err)
	}

	value, err := await(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("request Windows media session manager: %w", err)
	}
	if value == nil {
		return nil, fmt.Errorf("request Windows media session manager: empty result")
	}

	unknown := (*ole.IUnknown)(value)
	defer unknown.Release()

	var manager *control.GlobalSystemMediaTransportControlsSessionManager
	if err := unknown.PutQueryInterface(ole.NewGUID(control.GUIDiGlobalSystemMediaTransportControlsSessionManager), &manager); err != nil {
		return nil, fmt.Errorf("open Windows media session manager: %w", err)
	}
	return manager, nil
}

func safeCurrent(ctx context.Context, manager *control.GlobalSystemMediaTransportControlsSessionManager) (track media.Track, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			track = media.Track{Status: media.StatusUnknown}
			err = fmt.Errorf("read Windows media session: %v", recovered)
		}
	}()
	return current(ctx, manager)
}

func current(ctx context.Context, manager *control.GlobalSystemMediaTransportControlsSessionManager) (media.Track, error) {
	session, err := manager.GetCurrentSession()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("get active Windows media session: %w", err)
	}
	if session == nil {
		return trackFromSnapshot(snapshot{}, time.Now()), nil
	}
	defer session.Release()

	playback, err := session.GetPlaybackInfo()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback state: %w", err)
	}
	if playback == nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback state: empty result")
	}
	defer playback.Release()

	playbackValue, err := playback.GetPlaybackStatus()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback status: %w", err)
	}

	timeline, err := session.GetTimelineProperties()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback timeline: %w", err)
	}
	if timeline == nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback timeline: empty result")
	}
	defer timeline.Release()

	position, err := timeline.GetPosition()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows playback position: %w", err)
	}
	updated, err := timeline.GetLastUpdatedTime()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows timeline update time: %w", err)
	}

	propertiesOperation, err := session.TryGetMediaPropertiesAsync()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("request Windows media properties: %w", err)
	}
	propertiesValue, err := await(ctx, propertiesOperation)
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("request Windows media properties: %w", err)
	}
	if propertiesValue == nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("request Windows media properties: empty result")
	}

	unknown := (*ole.IUnknown)(propertiesValue)
	defer unknown.Release()

	var properties *control.GlobalSystemMediaTransportControlsSessionMediaProperties
	if err := unknown.PutQueryInterface(ole.NewGUID(control.GUIDiGlobalSystemMediaTransportControlsSessionMediaProperties), &properties); err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("open Windows media properties: %w", err)
	}
	defer properties.Release()

	title, err := properties.GetTitle()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows media title: %w", err)
	}
	artist, err := properties.GetArtist()
	if err != nil {
		return media.Track{Status: media.StatusUnknown}, fmt.Errorf("read Windows media artist: %w", err)
	}

	return trackFromSnapshot(snapshot{
		available:        true,
		title:            title,
		artist:           artist,
		positionTicks:    position.Duration,
		lastUpdatedTicks: updated.UniversalTime,
		playbackStatus:   int32(playbackValue),
	}, time.Now()), nil
}

func await(ctx context.Context, operation *foundation.IAsyncOperation) (unsafe.Pointer, error) {
	if operation == nil {
		return nil, fmt.Errorf("empty async operation")
	}
	defer operation.Release()

	var info *foundation.IAsyncInfo
	if err := operation.PutQueryInterface(ole.NewGUID(foundation.GUIDIAsyncInfo), &info); err != nil {
		return nil, fmt.Errorf("inspect async operation: %w", err)
	}
	defer info.Release()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		state, err := info.GetStatus()
		if err != nil {
			return nil, fmt.Errorf("read async operation status: %w", err)
		}

		switch state {
		case foundation.AsyncStatusCompleted:
			return operation.GetResults()
		case foundation.AsyncStatusCanceled:
			return nil, context.Canceled
		case foundation.AsyncStatusError:
			code, codeErr := info.GetErrorCode()
			if codeErr != nil {
				return nil, fmt.Errorf("read async operation error: %w", codeErr)
			}
			return nil, fmt.Errorf("Windows async operation failed with HRESULT 0x%08x", uint32(code.Value))
		}

		select {
		case <-ctx.Done():
			_ = info.Cancel()
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
