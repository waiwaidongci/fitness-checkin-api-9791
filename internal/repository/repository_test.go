package repository

import (
	"errors"
	"path/filepath"
	"testing"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/model"
)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()
	repo, err := Open(config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Close(); err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})
	return repo
}

func TestCreateAndGetWorkout(t *testing.T) {
	repo := newTestRepository(t)
	created, err := repo.Create(model.Workout{
		SportType:       "running",
		DurationMinutes: 30,
		Calories:        300,
		WorkoutDate:     "2026-08-15",
		Note:            "morning run",
	})
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero workout id")
	}

	got, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("get workout: %v", err)
	}
	if got.SportType != "running" || got.DurationMinutes != 30 || got.Calories != 300 {
		t.Fatalf("unexpected workout: %#v", got)
	}
}

func TestUpdateAndDeleteWorkout(t *testing.T) {
	repo := newTestRepository(t)
	created, err := repo.Create(model.Workout{
		SportType:       "walking",
		DurationMinutes: 20,
		Calories:        100,
		WorkoutDate:     "2026-08-14",
		Note:            "walk",
	})
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}

	created.SportType = "cycling"
	created.DurationMinutes = 45
	created.Calories = 360
	updated, err := repo.Update(created)
	if err != nil {
		t.Fatalf("update workout: %v", err)
	}
	if updated.SportType != "cycling" || updated.DurationMinutes != 45 {
		t.Fatalf("unexpected updated workout: %#v", updated)
	}

	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("delete workout: %v", err)
	}
	_, err = repo.GetByID(created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListFiltersAndPagination(t *testing.T) {
	repo := newTestRepository(t)
	workouts := []model.Workout{
		{SportType: "running", DurationMinutes: 10, Calories: 100, WorkoutDate: "2026-08-10"},
		{SportType: "running", DurationMinutes: 20, Calories: 200, WorkoutDate: "2026-08-11"},
		{SportType: "cycling", DurationMinutes: 30, Calories: 240, WorkoutDate: "2026-08-12"},
	}
	for _, w := range workouts {
		if _, err := repo.Create(w); err != nil {
			t.Fatalf("create workout: %v", err)
		}
	}

	items, total, err := repo.List(ListFilter{SportType: "running"})
	if err != nil {
		t.Fatalf("list running workouts: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 running workouts, got total=%d len=%d", total, len(items))
	}

	items, total, err = repo.List(ListFilter{
		StartDate: "2026-08-10",
		EndDate:   "2026-08-11",
		Page:      1,
		PageSize:  1,
	})
	if err != nil {
		t.Fatalf("list date range: %v", err)
	}
	if total != 2 || len(items) != 1 {
		t.Fatalf("expected total=2 and one page item, got total=%d len=%d", total, len(items))
	}
}

func TestAggregateBySportType(t *testing.T) {
	repo := newTestRepository(t)
	workouts := []model.Workout{
		{SportType: "running", DurationMinutes: 10, Calories: 100, WorkoutDate: "2026-08-10"},
		{SportType: "running", DurationMinutes: 20, Calories: 200, WorkoutDate: "2026-08-11"},
		{SportType: "cycling", DurationMinutes: 30, Calories: 240, WorkoutDate: "2026-08-12"},
	}
	for _, w := range workouts {
		if _, err := repo.Create(w); err != nil {
			t.Fatalf("create workout: %v", err)
		}
	}

	summaries, err := repo.AggregateBySportType("", "")
	if err != nil {
		t.Fatalf("aggregate workouts: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	bySport := make(map[string]model.SportSummary, len(summaries))
	for _, summary := range summaries {
		bySport[summary.SportType] = summary
	}
	if bySport["running"].TotalDuration != 30 || bySport["running"].TotalCalories != 300 {
		t.Fatalf("unexpected running summary: %#v", bySport["running"])
	}
	if bySport["cycling"].TotalDuration != 30 || bySport["cycling"].WorkoutCount != 1 {
		t.Fatalf("unexpected cycling summary: %#v", bySport["cycling"])
	}
}
