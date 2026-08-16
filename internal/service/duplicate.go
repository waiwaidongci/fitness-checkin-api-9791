package service

import (
	"errors"

	"fitness-checkin-api/internal/model"
	"fitness-checkin-api/internal/repository"
)

func (s *Service) duplicateExists(workout model.Workout) (bool, error) {
	if workout.Note == "" {
		return false, nil
	}

	_, err := s.repo.FindByUniqueKey(workout.SportType, workout.WorkoutDate, workout.Note)
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
