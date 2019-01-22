package wsHub

import (
	"log"
	"strconv"
	"github.com/gorilla/websocket"
	"github.com/gorilla/sessions"
	"net/http"
	"time"
	"encoding/json"
	"spotifyGroupQueueGo/userStore"
)

type Client struct {
	hub *Hub
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{}

func (client *Client) writer(roomCode string, store *userStore.Store, id string) {
	// Heroku timesout connections after 55 seconds (https://devcenter.heroku.com/articles/http-routing#timeouts)
	ticker := time.NewTicker(50 * time.Second)

	for {
		select {
			// Block until there is a message
			case message := <-client.send:
				log.Println("ws writing message")
				err := client.conn.WriteMessage(websocket.TextMessage, message)

				if err != nil {
					log.Println(err)
				}
			// Ping the client to see if they're still there
			case <-ticker.C:
				if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("ticker", err)
					ticker.Stop()
					client.conn.Close()
					client.hub.removeConnection(client, roomCode)

					// Only remove the user if it is their active websocket connection
					if store.IsActiveConn(roomCode, client.conn) {
						// We don't want to count them in the total user count (for veting)
						store.RemoveUser(id, roomCode)
						userCount := strconv.Itoa(store.GetTotalUserCount(roomCode))

						// Update the front end
						msg := map[string]string { "type": "totalUserCountUpdate", "count": userCount }
						j, _ := json.Marshal(msg)

						client.hub.Broadcast(j, roomCode)
					}
				}
		}
	}
}

func WsHandler(hub *Hub, cStore *sessions.CookieStore, uStore *userStore.Store,  w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}
	
	// Need the room code to send out messages based on room
	q := r.URL.Query()
	roomCode := q.Get("roomCode")

	session, _ := cStore.Get(r, "groupQueue")
	id, _ := session.Values["id"].(string)

	// Add this user to the store
	if !uStore.UserExists(id, roomCode) {
		uStore.AddUser(id, roomCode, conn)
		userCount := strconv.Itoa(uStore.GetTotalUserCount(roomCode))

		// Update the front end
		msg := map[string]string { "type": "totalUserCountUpdate", "count": userCount }
		j, _ := json.Marshal(msg)

		hub.Broadcast(j, roomCode)
	} else {
		uStore.UpdateUserConn(id, roomCode, conn)
	}
	
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 512)}
	client.hub.addConnection(client, roomCode)
	go client.writer(roomCode, uStore, id)
}
