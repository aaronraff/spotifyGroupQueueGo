package wsHub

import (
	"log"
	"github.com/gorilla/websocket"
	"net/http"
)

type Client struct {
	hub *hub
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{}

func (client *Client) writer() {
	for {
		// Block until there is a message
		message := <-client.send
		client.conn.WriteMessage(websocket.TextMessage, message)
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
	go client.writer()
}
