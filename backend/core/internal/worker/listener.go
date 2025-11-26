package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/AmoghRao21/swarm-os/core/internal/events"
	"github.com/AmoghRao21/swarm-os/core/internal/models"
	"github.com/AmoghRao21/swarm-os/core/internal/ws"
	"gorm.io/gorm"
)

type JobWorker struct {
	db           *gorm.DB
	eventManager *events.EventManager
	wsHub        *ws.Hub // Reference to the WebSocket Hub
}

type JobUpdatePayload struct {
	JobID  string                 `json:"job_id"`
	Status string                 `json:"status"`
	Result map[string]interface{} `json:"result"`
}

// NewJobWorker initializes the worker with DB, Redis, and WebSocket dependencies
func NewJobWorker(db *gorm.DB, em *events.EventManager, hub *ws.Hub) *JobWorker {
	return &JobWorker{
		db:           db,
		eventManager: em,
		wsHub:        hub,
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

	// 1. Update Database (Persistence)
	resultBytes, _ := json.Marshal(payload.Result)

	// We use Updates to only change specific fields
	result := w.db.Model(&models.Job{}).
		Where("id = ?", payload.JobID).
		Updates(map[string]interface{}{
			"status": payload.Status,
			"result": string(resultBytes),
		})

	if result.Error != nil {
		log.Printf("❌ Failed to update DB for Job [%s]: %v", payload.JobID, result.Error)
	} else {
		log.Printf("💾 Database updated: Job [%s] -> %s", payload.JobID, payload.Status)
	}

	// 2. Broadcast to WebSocket (Real-Time UI Update)
	// We structure this message specifically for the Frontend to consume
	updateMsg := map[string]interface{}{
		"type":   "JOB_UPDATE",
		"job_id": payload.JobID,
		"status": payload.Status,
		"data":   payload.Result,
	}

	w.wsHub.BroadcastToClients(updateMsg)
	log.Printf("📡 WebSocket broadcast sent for Job [%s]", payload.JobID)
}
