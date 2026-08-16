package cache

import (
	"sync"
	"time"

	"fitness-checkin-api/internal/model"
)

// RecentWeek keeps the last computed seven-day workout list in memory.
// The service refreshes it from a goroutine while HTTP handlers read it.
type RecentWeek struct {
	mu          sync.RWMutex
	items       []model.Workout
	refreshedAt time.Time
}

func NewRecentWeek() *RecentWeek {
	return &RecentWeek{}
}

func (c *RecentWeek) Replace(items []model.Workout) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = model.CloneWorkouts(items)
	c.refreshedAt = time.Now()
}

func (c *RecentWeek) Snapshot() []model.Workout {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return model.CloneWorkouts(c.items)
}

func (c *RecentWeek) IsFresh(maxAge time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return !c.refreshedAt.IsZero() && time.Since(c.refreshedAt) < maxAge
}

func (c *RecentWeek) RefreshedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.refreshedAt
}
