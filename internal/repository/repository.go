package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/model"
	"fitness-checkin-api/migrations"
)

var ErrNotFound = errors.New("workout not found")

type Repository struct {
	db *sql.DB
}

func Open(cfg config.Config) (*Repository, error) {
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite handles concurrent readers, but a single writer connection keeps
	// write behavior predictable for this small API.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	repo := &Repository{db: db}
	if err := repo.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) migrate() error {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := fs.ReadFile(migrations.FS, entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		for _, statement := range splitStatements(string(content)) {
			if _, err := r.db.Exec(statement); err != nil {
				return fmt.Errorf("run migration %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

func splitStatements(content string) []string {
	raw := strings.Split(content, ";")
	statements := make([]string, 0, len(raw))
	for _, part := range raw {
		statement := strings.TrimSpace(part)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}

func (r *Repository) Create(w model.Workout) (model.Workout, error) {
	now := time.Now().UTC().Truncate(time.Second)
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = now
	}

	result, err := r.db.Exec(
		`INSERT INTO workouts (sport_type, duration_minutes, calories, workout_date, note, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		w.SportType,
		w.DurationMinutes,
		w.Calories,
		w.WorkoutDate,
		w.Note,
		w.CreatedAt,
		w.UpdatedAt,
	)
	if err != nil {
		return model.Workout{}, fmt.Errorf("insert workout: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Workout{}, fmt.Errorf("get inserted workout id: %w", err)
	}
	w.ID = id
	return w, nil
}

func (r *Repository) GetByID(id int64) (model.Workout, error) {
	row := r.db.QueryRow(
		`SELECT id, sport_type, duration_minutes, calories, workout_date, note, created_at, updated_at
		 FROM workouts WHERE id = ?`,
		id,
	)
	return scanWorkout(row)
}

func (r *Repository) Update(w model.Workout) (model.Workout, error) {
	_, err := r.GetByID(w.ID)
	if err != nil {
		return model.Workout{}, err
	}

	w.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	_, err = r.db.Exec(
		`UPDATE workouts
		 SET sport_type = ?, duration_minutes = ?, calories = ?, workout_date = ?, note = ?, updated_at = ?
		 WHERE id = ?`,
		w.SportType,
		w.DurationMinutes,
		w.Calories,
		w.WorkoutDate,
		w.Note,
		w.UpdatedAt,
		w.ID,
	)
	if err != nil {
		return model.Workout{}, fmt.Errorf("update workout: %w", err)
	}
	return r.GetByID(w.ID)
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM workouts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete workout: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted row count: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type ListFilter struct {
	SportType string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

func (r *Repository) List(filter ListFilter) ([]model.Workout, int64, error) {
	return r.ListContext(context.Background(), filter)
}

func (r *Repository) ListContext(ctx context.Context, filter ListFilter) ([]model.Workout, int64, error) {
	where, args := buildWhere(filter.SportType, filter.StartDate, filter.EndDate)

	var total int64
	countQuery := "SELECT COUNT(*) FROM workouts" + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count workouts: %w", err)
	}

	query := `SELECT id, sport_type, duration_minutes, calories, workout_date, note, created_at, updated_at
	          FROM workouts` + where +
		` ORDER BY workout_date DESC, id DESC`
	if filter.PageSize > 0 {
		query += " LIMIT ? OFFSET ?"
		offset := (filter.Page - 1) * filter.PageSize
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list workouts: %w", err)
	}
	defer rows.Close()

	items := make([]model.Workout, 0)
	for rows.Next() {
		w, err := scanWorkoutRows(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, w)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate workouts: %w", err)
	}
	return items, total, nil
}

func (r *Repository) ListByDateRange(startDate, endDate string) ([]model.Workout, error) {
	items, _, err := r.List(ListFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PageSize:  0,
	})
	return items, err
}

func (r *Repository) AggregateBySportType(startDate, endDate string) ([]model.SportSummary, error) {
	where, args := buildWhere("", startDate, endDate)
	query := `SELECT sport_type, SUM(duration_minutes), SUM(calories), COUNT(*)
	          FROM workouts` + where +
		` GROUP BY sport_type
	          ORDER BY SUM(duration_minutes) DESC, sport_type ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("aggregate workouts: %w", err)
	}
	defer rows.Close()

	summaries := make([]model.SportSummary, 0)
	for rows.Next() {
		var summary model.SportSummary
		if err := rows.Scan(
			&summary.SportType,
			&summary.TotalDuration,
			&summary.TotalCalories,
			&summary.WorkoutCount,
		); err != nil {
			return nil, fmt.Errorf("scan workout summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout summaries: %w", err)
	}
	return summaries, nil
}

func buildWhere(sportType, startDate, endDate string) (string, []any) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if sportType != "" {
		conditions = append(conditions, "sport_type = ?")
		args = append(args, sportType)
	}
	if startDate != "" {
		conditions = append(conditions, "workout_date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conditions = append(conditions, "workout_date <= ?")
		args = append(args, endDate)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWorkout(row scanner) (model.Workout, error) {
	var w model.Workout
	err := row.Scan(
		&w.ID,
		&w.SportType,
		&w.DurationMinutes,
		&w.Calories,
		&w.WorkoutDate,
		&w.Note,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Workout{}, ErrNotFound
	}
	if err != nil {
		return model.Workout{}, fmt.Errorf("scan workout: %w", err)
	}
	return w, nil
}

func scanWorkoutRows(rows *sql.Rows) (model.Workout, error) {
	var w model.Workout
	err := rows.Scan(
		&w.ID,
		&w.SportType,
		&w.DurationMinutes,
		&w.Calories,
		&w.WorkoutDate,
		&w.Note,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return model.Workout{}, fmt.Errorf("scan workout row: %w", err)
	}
	return w, nil
}
