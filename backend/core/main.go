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

	// 3. Start Background Worker (The Listener)
	bgWorker := worker.NewJobWorker(db, eventManager)
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Run worker in a goroutine so it doesn't block the API
	go bgWorker.Start(ctx)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			sqlDB, _ := db.DB()
			err := sqlDB.Ping()
			status := "operational"
			if err != nil {
				status = "degraded"
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    status,
				"service":   "swarm-core",
				"timestamp": time.Now().Unix(),
			})
		})

		api.POST("/job", func(c *gin.Context) {
			var request struct {
				Task string `json:"task" binding:"required"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "task is required"})
				return
			}

			// A. Create Job in Postgres
			job := models.Job{
				Task:   request.Task,
				Status: models.JobStatusQueued,
				Result: "{}",
			}

			if result := db.Create(&job); result.Error != nil {
				log.Printf("Database Error: %v", result.Error)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist job"})
				return
			}

			// B. Prepare Payload for Brain
			payloadData := map[string]interface{}{
				"job_id": job.ID.String(),
				"task":   job.Task,
			}
			payload, _ := json.Marshal(payloadData)

			// C. Dispatch to Redis
			go func() {
				pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := eventManager.PublishEvent(pubCtx, "job_queue", payload); err != nil {
					log.Printf("Error publishing job %s: %v", job.ID, err)
				}
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

	cancelCtx() // Stop the worker gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}
