package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/s4kh/backend/models"
)

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan *models.Vote
	mutex     sync.RWMutex
	upgrader  websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *models.Vote),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) HandleWebSocket() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("failed to upgrade to ws connection: %v", err)
			return
		}
		defer conn.Close()

		h.mutex.Lock()
		h.clients[conn] = true
		h.mutex.Unlock()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("connection closed: %v", err)
				h.mutex.Lock()
				delete(h.clients, conn)
				h.mutex.Unlock()
				return
			}
		}

	})
}

func (h *Hub) BroadcastVoteUpdate(v *models.Vote) {
	h.broadcast <- v
}

func (h *Hub) StartBroadCast() {
	for {
		vote := <-h.broadcast
		h.mutex.Lock()
		for client := range h.clients {
			err := client.WriteJSON(vote)
			if err != nil {
				log.Printf("failed to send msg to client: %v", err)
				client.Close()
				delete(h.clients, client)
			}
		}
		h.mutex.Unlock()
	}
}
