package handler_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestWorkoutSummaryEndpoint(t *testing.T) {
	server := newTestServer(t)
	date := time.Now().Format("2006-01-02")
	created := createWorkout(t, server.URL, "swimming", 20, date)

	response, err := http.Get(server.URL + "/api/v1/workouts/" + strconv.FormatInt(created.ID, 10) + "/summary")
	if err != nil {
		t.Fatalf("get workout summary: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}

	var payload struct {
		TotalDurationMinutes int     `json:"total_duration_minutes"`
		TotalCalories        float64 `json:"total_calories"`
		WorkoutCount         int     `json:"workout_count"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode workout summary: %v", err)
	}
	if payload.TotalDurationMinutes != 20 || payload.TotalCalories != 180 || payload.WorkoutCount != 1 {
		t.Fatalf("unexpected summary: %#v", payload)
	}
}
