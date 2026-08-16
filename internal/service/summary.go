package service

import (
	"fmt"
	"math"
	"strings"

	"fitness-checkin-api/internal/model"
)

func (s *Service) buildWorkoutSummary(workout model.Workout) (*model.SportSummary, error) {
	if workout.ID <= 0 {
		return nil, fmt.Errorf("%w: workout id is required", ErrInvalidInput)
	}

	sportType := normalizeSportType(workout.SportType)
	if sportType == "" {
		return nil, fmt.Errorf("%w: sport_type is required", ErrInvalidInput)
	}
	if workout.DurationMinutes <= 0 {
		return nil, fmt.Errorf("%w: duration_minutes must be greater than 0", ErrInvalidInput)
	}
	if !validDate(workout.WorkoutDate) {
		return nil, fmt.Errorf("%w: workout_date must use YYYY-MM-DD", ErrInvalidInput)
	}

	calories := s.CaloriesFor(sportType, workout.DurationMinutes)
	return &model.SportSummary{
		SportType:     sportType,
		TotalDuration: workout.DurationMinutes,
		TotalCalories: math.Round(calories*10) / 10,
		WorkoutCount:  1,
	}, nil
}

func summarizeWorkout(sportType string, durationMinutes int, calories float64) model.SportSummary {
	return model.SportSummary{
		SportType:     strings.ToLower(strings.TrimSpace(sportType)),
		TotalDuration: durationMinutes,
		TotalCalories: math.Round(calories*10) / 10,
		WorkoutCount:  1,
	}
}
