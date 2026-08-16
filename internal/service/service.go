package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"fitness-checkin-api/internal/cache"
	"fitness-checkin-api/internal/model"
	"fitness-checkin-api/internal/repository"
)

var (
	ErrNotFound     = errors.New("workout not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInvalidRange = errors.New("invalid date range")
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

var calorieRates = map[string]float64{
	"running":  10.0,
	"cycling":  8.0,
	"swimming": 9.0,
	"walking":  5.0,
	"yoga":     4.0,
	"hiking":   7.0,
	"strength": 6.0,
}

const defaultCalorieRate = 6.0

type WorkoutRepository interface {
	Create(model.Workout) (model.Workout, error)
	GetByID(int64) (model.Workout, error)
	Update(model.Workout) (model.Workout, error)
	Delete(int64) error
	List(repository.ListFilter) ([]model.Workout, int64, error)
	ListByDateRange(string, string) ([]model.Workout, error)
	AggregateBySportType(string, string) ([]model.SportSummary, error)
}

type Service struct {
	repo      WorkoutRepository
	now       func() time.Time
	recent    *cache.RecentWeek
	freshness cache.FreshnessPolicy
	stop      chan struct{}
	startOnce sync.Once
}

func New(repo WorkoutRepository) *Service {
	return &Service{
		repo:      repo,
		now:       time.Now,
		recent:    cache.NewRecentWeek(),
		freshness: cache.DefaultFreshnessPolicy(),
		stop:      make(chan struct{}),
	}
}

type WorkoutInput struct {
	SportType       string
	DurationMinutes int
	WorkoutDate     string
	Note            string
}

type ListParams struct {
	SportType string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

type PaginatedWorkouts struct {
	Items    []model.Workout `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func (s *Service) Create(input WorkoutInput) (model.Workout, error) {
	workout, err := s.prepareWorkout(0, input)
	if err != nil {
		return model.Workout{}, err
	}
	return s.repo.Create(workout)
}

func (s *Service) GetByID(id int64) (model.Workout, error) {
	workout, err := s.repo.GetByID(id)
	if err != nil {
		return model.Workout{}, mapRepositoryError(err)
	}
	return workout, nil
}

func (s *Service) Update(id int64, input WorkoutInput) (model.Workout, error) {
	if id <= 0 {
		return model.Workout{}, ErrInvalidInput
	}
	workout, err := s.prepareWorkout(id, input)
	if err != nil {
		return model.Workout{}, err
	}
	updated, err := s.repo.Update(workout)
	if err != nil {
		return model.Workout{}, mapRepositoryError(err)
	}
	return updated, nil
}

func (s *Service) Delete(id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	return mapRepositoryError(s.repo.Delete(id))
}

func (s *Service) List(params ListParams) (PaginatedWorkouts, error) {
	if err := validateRange(params.StartDate, params.EndDate); err != nil {
		return PaginatedWorkouts{}, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	items, total, err := s.repo.List(repository.ListFilter{
		SportType: normalizeSportType(params.SportType),
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return PaginatedWorkouts{}, err
	}
	return PaginatedWorkouts{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) RecentWeek() ([]model.Workout, error) {
	if s.freshness.Allows(s.recent.RefreshedAt(), s.now()) {
		return s.recent.Snapshot(), nil
	}
	return s.refreshRecentWeekNow()
}

func (s *Service) refreshRecentWeekNow() ([]model.Workout, error) {
	today := s.now().Format(model.DateLayout)
	start := s.now().AddDate(0, 0, -6).Format(model.DateLayout)
	items, err := s.repo.ListByDateRange(start, today)
	if err != nil {
		return nil, err
	}
	s.recent.Replace(items)
	return s.recent.Snapshot(), nil
}

func (s *Service) StartCacheRefresh() {
	s.startOnce.Do(func() {
		go s.refreshRecentWeek()
	})
}

func (s *Service) Close() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
}

func (s *Service) refreshRecentWeek() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			_, _ = s.refreshRecentWeekNow()
		}
	}
}

func (s *Service) SummaryBySportType(startDate, endDate string) ([]model.SportSummary, error) {
	if err := validateRange(startDate, endDate); err != nil {
		return nil, err
	}
	return s.repo.AggregateBySportType(startDate, endDate)
}

func (s *Service) CaloriesFor(sportType string, durationMinutes int) float64 {
	rate, ok := calorieRates[normalizeSportType(sportType)]
	if !ok {
		rate = defaultCalorieRate
	}
	calories := rate * float64(durationMinutes)
	return math.Round(calories*10) / 10
}

func (s *Service) prepareWorkout(id int64, input WorkoutInput) (model.Workout, error) {
	sportType := normalizeSportType(input.SportType)
	if sportType == "" {
		return model.Workout{}, fmt.Errorf("%w: sport_type is required", ErrInvalidInput)
	}
	if input.DurationMinutes <= 0 {
		return model.Workout{}, fmt.Errorf("%w: duration_minutes must be greater than 0", ErrInvalidInput)
	}
	if !validDate(input.WorkoutDate) {
		return model.Workout{}, fmt.Errorf("%w: workout_date must use YYYY-MM-DD", ErrInvalidInput)
	}

	return model.Workout{
		ID:              id,
		SportType:       sportType,
		DurationMinutes: input.DurationMinutes,
		Calories:        s.CaloriesFor(sportType, input.DurationMinutes),
		WorkoutDate:     input.WorkoutDate,
		Note:            strings.TrimSpace(input.Note),
	}, nil
}

func normalizeSportType(sportType string) string {
	return strings.ToLower(strings.TrimSpace(sportType))
}

func validDate(value string) bool {
	parsed, err := time.Parse(model.DateLayout, value)
	if err != nil {
		return false
	}
	return parsed.Format(model.DateLayout) == value
}

func validateRange(startDate, endDate string) error {
	if startDate != "" && !validDate(startDate) {
		return fmt.Errorf("%w: start_date must use YYYY-MM-DD", ErrInvalidInput)
	}
	if endDate != "" && !validDate(endDate) {
		return fmt.Errorf("%w: end_date must use YYYY-MM-DD", ErrInvalidInput)
	}
	if startDate != "" && endDate != "" && startDate > endDate {
		return ErrInvalidRange
	}
	return nil
}

func mapRepositoryError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
