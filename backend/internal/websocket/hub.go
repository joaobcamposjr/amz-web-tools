package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub mantém as conexões WebSocket ativas
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// Client representa uma conexão WebSocket
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// LogMessage representa uma mensagem de log
type LogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Step      string `json:"step"`
	Message   string `json:"message"`
	ProcessID string `json:"process_id,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permitir todas as origens para desenvolvimento
	},
}

// NewHub cria uma nova instância do Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run inicia o hub WebSocket
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("🔌 Cliente WebSocket conectado. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("🔌 Cliente WebSocket desconectado. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastLog envia um log para todos os clientes conectados
func (h *Hub) BroadcastLog(logMsg LogMessage) {
	message := map[string]interface{}{
		"type":       logMsg.Type,
		"timestamp":  logMsg.Timestamp,
		"level":      logMsg.Level,
		"step":       logMsg.Step,
		"message":    logMsg.Message,
		"process_id": logMsg.ProcessID,
	}

	// Converter para JSON e enviar
	if jsonData, err := json.Marshal(message); err == nil {
		h.broadcast <- jsonData
	}
}

// GetClientCount retorna o número de clientes conectados
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
