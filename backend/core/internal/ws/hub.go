package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (CORS)
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Println("🔌 Client connected to WebSocket")

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
			log.Println("🔌 Client disconnected")

		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				// FIX: Removed invalid 'select' block.
				// Write directly to the client.
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("❌ WebSocket Write Error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// BroadcastToClients sends a message to all connected UIs
func (h *Hub) BroadcastToClients(message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling broadcast: %v", err)
		return
	}
	h.broadcast <- bytes
}

// HandleWS upgrades HTTP requests to WebSocket connections
func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WS: %v", err)
		return
	}
	h.register <- conn
}
