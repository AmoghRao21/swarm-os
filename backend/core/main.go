package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarmos/core/internal/events"
)

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}

	eventManager, err := events.NewEventManager(redisUrl)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize event infrastructure: %v", err)
	}
	defer eventManager.Close()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "operational",
				"service":   "swarm-core",
				"timestamp": time.Now().Unix(),
			})
		})

		api.POST("/job", func(c *gin.Context) {
			var jobData map[string]interface{}
			if err := c.BindJSON(&jobData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
				return
			}
			
			// Fire and forget to the Brain
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := eventManager.PublishEvent(ctx, "job_queue", jobData); err != nil {
					log.Printf("Error publishing job: %v", err)
				}
			}()

			c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
		})
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}