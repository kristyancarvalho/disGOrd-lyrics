package status

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	Update(context.Context, string) error
	Clear(context.Context) error
}

type Manager struct {
	client      Client
	mu          sync.Mutex
	last        string
	initialized bool
	attempted   string
	retryAt     time.Time
	retryDelay  time.Duration
	now         func() time.Time
}

func NewManager(client Client) *Manager {
	return &Manager{
		client:     client,
		retryDelay: 5 * time.Second,
		now:        time.Now,
	}
}

func (manager *Manager) Set(ctx context.Context, text string) (bool, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.initialized && manager.last == text {
		return false, nil
	}

	now := manager.now()
	if manager.attempted == text && now.Before(manager.retryAt) {
		return false, nil
	}

	if text == "" {
		if err := manager.client.Clear(ctx); err != nil {
			manager.attempted = text
			manager.retryAt = now.Add(manager.retryDelay)
			return false, err
		}
	} else {
		if err := manager.client.Update(ctx, text); err != nil {
			manager.attempted = text
			manager.retryAt = now.Add(manager.retryDelay)
			return false, err
		}
	}

	manager.last = text
	manager.initialized = true
	manager.attempted = ""
	manager.retryAt = time.Time{}
	return true, nil
}

func (manager *Manager) Clear(ctx context.Context) (bool, error) {
	return manager.Set(ctx, "")
}
