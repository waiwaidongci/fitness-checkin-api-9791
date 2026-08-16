package service

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/repository"
)

func TestRecentWeekConcurrent(t *testing.T) {
	repo, err := repository.Open(config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	defer repo.Close()

	svc := New(repo)
	svc.now = func() time.Time {
		return time.Date(2026, 8, 16, 12, 0, 0, 0, time.Local)
	}
	for i := 0; i < 4; i++ {
		if _, err := svc.Create(WorkoutInput{
			SportType:       "running",
			DurationMinutes: 10,
			WorkoutDate:     "2026-08-15",
		}); err != nil {
			t.Fatalf("create workout: %v", err)
		}
	}

	svc.StartCacheRefresh()
	defer svc.Close()

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := svc.RecentWeek(); err != nil {
				t.Errorf("recent week: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
}
