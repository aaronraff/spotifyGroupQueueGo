package wsHub

import (
	"log"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type Client struct {
	hub *hub
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{}

func (client *Client) writer(roomCode string) {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
			// Block until there is a message
			case message := <-client.send:
				client.conn.WriteMessage(websocket.TextMessage, message)
			// Ping the client to see if they're still there
			case <-ticker.C:
				if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					ticker.Stop()
					client.conn.Close()
					client.hub.removeConnection(client, roomCode)
				}
		}
	}
}

func WsHandler(hub *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}
	
	// Need the room code to send out messages based on room
	q := r.URL.Query()
	roomCode := q.Get("roomCode")
	log.Println(roomCode)
	
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 512)}
	client.hub.addConnection(client, roomCode)
	go client.writer(roomCode)
}
