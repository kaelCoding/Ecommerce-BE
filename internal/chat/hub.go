package chat

import (
	"log"
)

type Hub struct {
	clients map[uint]*Client
	broadcast chan []byte
	register chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[uint]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("Client registered: UserID %d", client.userID)
			h.clients[client.userID] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("Client unregistered: UserID %d", client.userID)
			}
		case message := <-h.broadcast:
			log.Printf("Message received in hub: %s", string(message))
		}
	}
}
