package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"fitness-checkin-api/internal/config"
	"fitness-checkin-api/internal/repository"
	"fitness-checkin-api/internal/router"
	"fitness-checkin-api/internal/service"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	repo, err := repository.Open(cfg)
	if err != nil {
		log.Fatalf("initialize repository: %v", err)
	}
	defer repo.Close()

	svc := service.New(repo)
	engine := router.Setup(svc)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("fitness check-in API listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
