package service

import (
	"errors"
	"fmt"
	"strings"

	"fitness-checkin-api/internal/model"
	"fitness-checkin-api/internal/repository"
)

type duplicatePolicy struct {
	repo WorkoutRepository
}

func newDuplicatePolicy(repo WorkoutRepository) duplicatePolicy {
	return duplicatePolicy{repo: repo}
}

func (p duplicatePolicy) reject(workout model.Workout) error {
	note := strings.TrimSpace(workout.Note)
	if note == "" {
		return nil
	}

	_, err := p.repo.FindByUniqueKey(workout.SportType, workout.WorkoutDate, note)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("%w: workout already exists", ErrDuplicateWorkout)
}
