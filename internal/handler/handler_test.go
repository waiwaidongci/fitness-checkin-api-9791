package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/model"
	"fitness-checkin-api/internal/repository"
	"fitness-checkin-api/internal/router"
	"fitness-checkin-api/internal/service"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	repo, err := repository.Open(config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	engine := router.Setup(service.New(repo))
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	return server
}

func TestWorkoutEndpoints(t *testing.T) {
	server := newTestServer(t)
	date := time.Now().Format("2006-01-02")

	running := createWorkout(t, server.URL, "running", 30, date)
	cycling := createWorkout(t, server.URL, "cycling", 45, date)
	if running.ID == 0 || cycling.ID == 0 {
		t.Fatalf("expected created ids, got running=%d cycling=%d", running.ID, cycling.ID)
	}

	list := getJSON[service.PaginatedWorkouts](t, server.URL+"/api/v1/workouts?start_date="+date+"&end_date="+date)
	if list.Total != 2 || len(list.Items) != 2 {
		t.Fatalf("expected two workouts in date range, got %#v", list)
	}

	summary := getJSON[map[string]any](t, server.URL+"/api/v1/workouts/summary/by-sport-type?start_date="+date+"&end_date="+date)
	items, ok := summary["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two summaries, got %#v", summary)
	}

	recent := getJSON[map[string]any](t, server.URL+"/api/v1/workouts/recent-week")
	recentItems, ok := recent["items"].([]any)
	if !ok || len(recentItems) != 2 {
		t.Fatalf("expected two recent workouts, got %#v", recent)
	}

	updated := putJSON[model.Workout](t, fmt.Sprintf("%s/api/v1/workouts/%d", server.URL, running.ID), map[string]any{
		"sport_type":       "swimming",
		"duration_minutes": 20,
		"workout_date":     date,
		"note":             "updated",
	})
	if updated.SportType != "swimming" || updated.Calories != 180 {
		t.Fatalf("unexpected updated workout: %#v", updated)
	}

	status := deleteWorkout(t, fmt.Sprintf("%s/api/v1/workouts/%d", server.URL, cycling.ID))
	if status != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", status)
	}

	afterDelete := getJSON[service.PaginatedWorkouts](t, server.URL+"/api/v1/workouts?start_date="+date+"&end_date="+date)
	if afterDelete.Total != 1 {
		t.Fatalf("expected one workout after delete, got %#v", afterDelete)
	}
}

func createWorkout(t *testing.T, baseURL, sportType string, duration int, date string) model.Workout {
	t.Helper()
	return postJSON[model.Workout](t, baseURL+"/api/v1/workouts", map[string]any{
		"sport_type":       sportType,
		"duration_minutes": duration,
		"workout_date":     date,
		"note":             "test",
	})
}

func postJSON[T any](t *testing.T, url string, body any) T {
	t.Helper()
	return requestJSON[T](t, http.MethodPost, url, body, http.StatusCreated)
}

func putJSON[T any](t *testing.T, url string, body any) T {
	t.Helper()
	return requestJSON[T](t, http.MethodPut, url, body, http.StatusOK)
}

func getJSON[T any](t *testing.T, url string) T {
	t.Helper()
	var result T
	response, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("GET %s returned %d: %s", url, response.StatusCode, string(body))
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode GET %s: %v", url, err)
	}
	return result
}

func requestJSON[T any](t *testing.T, method, url string, body any, expectedStatus int) T {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	request, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if response.StatusCode != expectedStatus {
		t.Fatalf("%s %s returned %d, want %d: %s", method, url, response.StatusCode, expectedStatus, string(responseBody))
	}
	var result T
	if err := json.Unmarshal(responseBody, &result); err != nil {
		t.Fatalf("decode %s %s response: %v", method, url, err)
	}
	return result
}

func deleteWorkout(t *testing.T, url string) int {
	t.Helper()
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("new delete request: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("DELETE %s: %v", url, err)
	}
	defer response.Body.Close()
	return response.StatusCode
}
