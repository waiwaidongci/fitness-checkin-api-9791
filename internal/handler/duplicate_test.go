package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"fitness-checkin-api/internal/model"
)

func TestDuplicateWorkoutReturnsConflict(t *testing.T) {
	server := newTestServer(t)
	date := time.Now().Format("2006-01-02")
	body := map[string]any{
		"sport_type":       "running",
		"duration_minutes": 30,
		"workout_date":     date,
		"note":             "same note",
	}

	postJSON[model.Workout](t, server.URL+"/api/v1/workouts", body)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/workouts", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("create duplicate: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", response.StatusCode)
	}
}
