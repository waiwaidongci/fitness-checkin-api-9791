package cache

import (
	"sync"

	"fitness-checkin-api/internal/model"
)

// RecentWeek keeps the last computed seven-day workout list in memory.
// The service refreshes it from a goroutine while HTTP handlers read it.
type RecentWeek struct {
	mu    sync.RWMutex
	items []model.Workout
}

func NewRecentWeek() *RecentWeek {
	return &RecentWeek{}
}

func (c *RecentWeek) Replace(items []model.Workout) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = items
}

func (c *RecentWeek) Snapshot() []model.Workout {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items
}
