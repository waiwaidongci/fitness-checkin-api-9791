package router

import (
	"github.com/gin-gonic/gin"

	"fitness-checkin-api/internal/handler"
	"fitness-checkin-api/internal/middleware"
	"fitness-checkin-api/internal/service"
)

func Setup(service *service.Service) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.RequestID(), middleware.CORS())

	h := handler.New(service)
	api := engine.Group("/api/v1")
	{
		api.GET("/workouts", h.List)
		api.GET("/workouts/recent-week", h.RecentWeek)
		api.GET("/workouts/summary/by-sport-type", h.SummaryBySportType)
		api.GET("/workouts/:id/summary", h.WorkoutSummary)
		api.GET("/workouts/:id", h.Get)
		api.POST("/workouts", h.Create)
		api.PUT("/workouts/:id", h.Update)
		api.DELETE("/workouts/:id", h.Delete)
	}

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return engine
}
