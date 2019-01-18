package wsHub

import (
	"log"
)

type hub struct {
	clients map[string]*Client
}

func NewHub() *hub {
	return &hub { clients: make(map[string]*Client) }
}

func (h *hub) addConnection(c *Client, roomCode string) {
	log.Println("Added client:", c)
	h.clients[roomCode] = c	
}

func (h *hub) removeConnection(c *Client, roomCode string) {
	log.Println("Removed client:", c)
	delete(h.clients, roomCode)
}

func (h *hub) Broadcast(msg []byte, roomCode string) {
	for r, c := range h.clients {
		if r == roomCode {
			c.send <- msg
		}
	}
}
