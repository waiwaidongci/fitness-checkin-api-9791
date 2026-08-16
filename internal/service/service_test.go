package service

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/repository"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	repo, err := repository.Open(config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Close(); err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})
	return New(repo)
}

func TestCaloriesFor(t *testing.T) {
	svc := New(nil)
	tests := []struct {
		sportType       string
		durationMinutes int
		want            float64
	}{
		{"Running", 30, 300},
		{"cycling", 45, 360},
		{"unknown", 20, 120},
	}
	for _, test := range tests {
		if got := svc.CaloriesFor(test.sportType, test.durationMinutes); got != test.want {
			t.Fatalf("CaloriesFor(%q, %d) = %v, want %v", test.sportType, test.durationMinutes, got, test.want)
		}
	}
}

func TestCreateCalculatesCalories(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(WorkoutInput{
		SportType:       " Running ",
		DurationMinutes: 30,
		WorkoutDate:     "2026-08-15",
		Note:            " morning ",
	})
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}
	if created.SportType != "running" {
		t.Fatalf("expected normalized sport type, got %q", created.SportType)
	}
	if created.Calories != 300 {
		t.Fatalf("expected 300 calories, got %v", created.Calories)
	}
	if created.Note != "morning" {
		t.Fatalf("expected trimmed note, got %q", created.Note)
	}
}

func TestRecentWeek(t *testing.T) {
	svc := newTestService(t)
	svc.now = func() time.Time {
		return time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	}
	for _, workout := range []struct {
		date     string
		duration int
		sport    string
	}{
		{"2026-08-14", 10, "running"},
		{"2026-08-09", 10, "running"},
		{"2026-08-08", 10, "running"},
	} {
		if _, err := svc.Create(WorkoutInput{SportType: workout.sport, DurationMinutes: workout.duration, WorkoutDate: workout.date}); err != nil {
			t.Fatalf("create workout: %v", err)
		}
	}

	items, err := svc.RecentWeek()
	if err != nil {
		t.Fatalf("recent week: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two recent workouts, got %d", len(items))
	}
}

func TestListPagination(t *testing.T) {
	svc := newTestService(t)
	for i := 1; i <= 12; i++ {
		_, err := svc.Create(WorkoutInput{
			SportType:       "walking",
			DurationMinutes: 10,
			WorkoutDate:     "2026-08-15",
		})
		if err != nil {
			t.Fatalf("create workout %d: %v", i, err)
		}
	}

	result, err := svc.List(ListParams{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("list workouts: %v", err)
	}
	if result.Total != 12 {
		t.Fatalf("expected total 12, got %d", result.Total)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 items on page 2, got %d", len(result.Items))
	}
	if result.Page != 2 || result.PageSize != 5 {
		t.Fatalf("unexpected pagination metadata: %#v", result)
	}
}

func TestInvalidInputAndRange(t *testing.T) {
	svc := New(nil)
	_, err := svc.Create(WorkoutInput{SportType: "running", DurationMinutes: 30, WorkoutDate: "2026/08/15"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.List(ListParams{StartDate: "2026-08-16", EndDate: "2026-08-15"})
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}
