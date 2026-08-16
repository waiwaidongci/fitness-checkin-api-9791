package model

import "time"

const DateLayout = "2006-01-02"

type Workout struct {
	ID              int64         `json:"id"`
	SportType       string        `json:"sport_type"`
	DurationMinutes int           `json:"duration_minutes"`
	Calories        float64       `json:"calories"`
	WorkoutDate     string        `json:"workout_date"`
	Note            string        `json:"note"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Summary         *SportSummary `json:"summary,omitempty"`
}

type SportSummary struct {
	SportType     string  `json:"sport_type"`
	TotalDuration int     `json:"total_duration_minutes"`
	TotalCalories float64 `json:"total_calories"`
	WorkoutCount  int     `json:"workout_count"`
}
