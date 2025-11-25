package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/AmoghRao21/swarm-os/core/internal/events"
	"github.com/AmoghRao21/swarm-os/core/internal/models"
	"gorm.io/gorm"
)

type JobWorker struct {
	db           *gorm.DB
	eventManager *events.EventManager
}

type JobUpdatePayload struct {
	JobID  string                 `json:"job_id"`
	Status string                 `json:"status"`
	Result map[string]interface{} `json:"result"`
}

func NewJobWorker(db *gorm.DB, em *events.EventManager) *JobWorker {
	return &JobWorker{
		db:           db,
		eventManager: em,
	}
}

func (w *JobWorker) Start(ctx context.Context) {
	log.Println("👂 Job Worker started. Listening for updates from Brain...")
	pubsub := w.eventManager.Subscribe(ctx, "job_updates")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			w.handleMessage(msg.Payload)
		case <-ctx.Done():
			log.Println("🛑 Job Worker shutting down...")
			return
		}
	}
}

func (w *JobWorker) handleMessage(payloadStr string) {
	var payload JobUpdatePayload
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		log.Printf("❌ Worker failed to parse message: %v", err)
		return
	}

	// Convert Result map to JSON string for Postgres storage
	resultBytes, _ := json.Marshal(payload.Result)

	// Update the Database
	result := w.db.Model(&models.Job{}).
		Where("id = ?", payload.JobID).
		Updates(map[string]interface{}{
			"status": payload.Status,
			"result": string(resultBytes),
		})

	if result.Error != nil {
		log.Printf("❌ Failed to update Job [%s]: %v", payload.JobID, result.Error)
	} else {
		log.Printf("💾 Database updated: Job [%s] -> %s", payload.JobID, payload.Status)
	}
}
