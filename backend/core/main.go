package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmoghRao21/swarm-os/core/internal/database"
	"github.com/AmoghRao21/swarm-os/core/internal/events"
	"github.com/AmoghRao21/swarm-os/core/internal/models"
	"github.com/AmoghRao21/swarm-os/core/internal/worker"
	"github.com/AmoghRao21/swarm-os/core/internal/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Redis
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}
	eventManager, err := events.NewEventManager(redisUrl)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize event infrastructure: %v", err)
	}
	defer eventManager.Close()

	// 2. Initialize Database
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "host=localhost user=swarm_admin password=secure_dev_password dbname=swarm_os port=5432 sslmode=disable"
	}
	db, err := database.NewDatabase(dbUrl)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize database: %v", err)
	}

	// 3. Initialize WebSocket Hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// 4. Start Background Worker
	bgWorker := worker.NewJobWorker(db, eventManager, wsHub)
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	go bgWorker.Start(ctx)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api := router.Group("/api/v1")
	{
		// WebSocket Endpoint
		api.GET("/ws", wsHub.HandleWS)

		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "operational"})
		})

		// NEW: Get Job Status (Hydration)
		api.GET("/job/:id", func(c *gin.Context) {
			id := c.Param("id")
			var job models.Job

			if err := db.First(&job, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
				return
			}

			// Parse the JSONB result string back into an object
			var resultData map[string]interface{}
			if job.Result != "" {
				json.Unmarshal([]byte(job.Result), &resultData)
			}

			c.JSON(http.StatusOK, gin.H{
				"job_id": job.ID,
				"status": job.Status,
				"data":   resultData,
			})
		})

		// Create Job
		api.POST("/job", func(c *gin.Context) {
			var request struct {
				Task    string `json:"task" binding:"required"`
				SwarmID string `json:"swarm_id" binding:"required"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "task and swarm_id are required"})
				return
			}

			job := models.Job{
				Task:   request.Task,
				Status: models.JobStatusQueued,
				Result: "{}",
			}

			if result := db.Create(&job); result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist job"})
				return
			}

			payloadData := map[string]interface{}{
				"job_id":   job.ID.String(),
				"task":     job.Task,
				"swarm_id": request.SwarmID,
			}
			payload, _ := json.Marshal(payloadData)

			go func() {
				pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				eventManager.PublishEvent(pubCtx, "job_queue", payload)
			}()

			c.JSON(http.StatusAccepted, gin.H{
				"status": "queued",
				"job_id": job.ID,
			})
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

	cancelCtx()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
