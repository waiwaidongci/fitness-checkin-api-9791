package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fitness-checkin-api/internal/service"
)

type WorkoutHandler struct {
	service *service.Service
}

func New(service *service.Service) *WorkoutHandler {
	return &WorkoutHandler{service: service}
}

type workoutRequest struct {
	SportType       string `json:"sport_type" binding:"required"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,gt=0"`
	WorkoutDate     string `json:"workout_date" binding:"required"`
	Note            string `json:"note"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *WorkoutHandler) Create(c *gin.Context) {
	var request workoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	workout, err := h.service.Create(service.WorkoutInput{
		SportType:       request.SportType,
		DurationMinutes: request.DurationMinutes,
		WorkoutDate:     request.WorkoutDate,
		Note:            request.Note,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, workout)
}

func (h *WorkoutHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	workout, err := h.service.GetByID(id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var request workoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	workout, err := h.service.Update(id, service.WorkoutInput{
		SportType:       request.SportType,
		DurationMinutes: request.DurationMinutes,
		WorkoutDate:     request.WorkoutDate,
		Note:            request.Note,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, workout)
}

func (h *WorkoutHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WorkoutHandler) List(c *gin.Context) {
	params := service.ListParams{
		SportType: c.Query("sport_type"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Page:      queryInt(c, "page", 0),
		PageSize:  queryInt(c, "page_size", 0),
	}
	result, err := h.service.ListContext(c.Request.Context(), params)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *WorkoutHandler) SummaryBySportType(c *gin.Context) {
	summaries, err := h.service.SummaryBySportType(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": summaries})
}

func (h *WorkoutHandler) RecentWeek(c *gin.Context) {
	workouts, err := h.service.RecentWeek()
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": workouts})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid workout id")
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(c, http.StatusNotFound, "workout not found")
	case errors.Is(err, service.ErrInvalidInput):
		respondError(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidRange):
		respondError(c, http.StatusBadRequest, "start_date must not be after end_date")
	default:
		respondError(c, http.StatusInternalServerError, "internal server error")
	}
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, errorResponse{Error: message})
}
