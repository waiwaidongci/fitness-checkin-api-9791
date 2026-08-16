package repository

import "fitness-checkin-api/internal/model"

func (r *Repository) FindByUniqueKey(sportType, workoutDate, note string) (model.Workout, error) {
	row := r.db.QueryRow(
		`SELECT id, sport_type, duration_minutes, calories, workout_date, note, created_at, updated_at
		 FROM workouts
		 WHERE sport_type = ? AND workout_date = ? AND note = ?
		 ORDER BY id DESC
		 LIMIT 1`,
		sportType,
		workoutDate,
		note,
	)
	return scanWorkout(row)
}
