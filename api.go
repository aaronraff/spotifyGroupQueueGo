package main

import (
	"log"
	"strconv"
	"net/http"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/zmb3/spotify"
	"spotifyGroupQueueGo/wsHub"
	"spotifyGroupQueueGo/workerStore"
	"spotifyGroupQueueGo/userStore"
	"spotifyGroupQueueGo/sessionStore"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func OpenRoomHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	client := auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	// Generate a code for the new room
	code := make([]byte, 7)
	rand.Read(code)

	// Need to make it base64
	str := base64.StdEncoding.EncodeToString(code)

	// Need to cut off at 7 chars (base64 can be longer)
	roomCode := str[:7]

	// Add the room to the DB
	InsertRoom(Db, roomCode, string(user.ID), tok)

	// Make sure there is atleast one song in the playlist
	checkAddRandomSong(&client, roomCode)
	
	// Update front end with room code for /room/start
	msg := map[string]string { "roomCode": roomCode }
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	// Success
	w.Header().Set("StatusCode", "200")
	w.Write(j)
}


// This will be hit after the user confirms that they have started the first 
// song in the playlist
func StartPollerHandler(hub *wsHub.Hub, uStore *userStore.Store, w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	roomCode := r.FormValue("roomCode")

	client := auth.NewClient(tok)
	notifyChan := UStore.AddChannel(roomCode)
	cancelChan := workerStore.AddPoller(roomCode)
	go PollPlayerForRemoval(&client, roomCode, hub, uStore, notifyChan, cancelChan)

	w.WriteHeader(200)
}

func CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	roomCode := r.FormValue("roomCode")

	client := auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	// Stop the poller
	cancelChan := workerStore.GetPollerChan(roomCode)
	cancelChan <- true
	workerStore.RemovePoller(roomCode)

	// Remove the room from the DB
	DeleteRoom(Db, string(user.ID))
	UStore.RemoveRoom(roomCode)

	msg := map[string]interface{} { "type": "roomClosed" }
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	WsHub.Broadcast(j, roomCode)

	// Success
	w.WriteHeader(200)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	songName := r.FormValue("songName")
	roomCode := r.FormValue("roomCode")

	// Get the token
	tok := GetTokenFromCode(Db, roomCode)

	client := auth.NewClient(tok)

	results, err := client.Search(songName, spotify.SearchTypeTrack)

	if err != nil {
		log.Println(err)
	}

	resJson, err := json.Marshal(results.Tracks.Tracks)

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")

	// Send response to client
	w.Write(resJson)
}

func AddToQueueHandler(w http.ResponseWriter, r *http.Request) {
	songID := r.FormValue("songID")
	roomCode := r.FormValue("roomCode")

	// Get the token
	tok := GetTokenFromCode(Db, roomCode)

	client := auth.NewClient(tok)
	groupPlaylistId := GetPlaylistIdByName(&client, "GroupQueue")

	tracks, err := client.GetPlaylistTracks(groupPlaylistId)

	if err != nil {
		log.Println(err)
	}

	if IsSongPresent(tracks.Tracks, songID) {
		msg := map[string]interface{} { "msg": "This song is already in the queue." }
		j, err := json.Marshal(msg)

		if err != nil {
			log.Println(err)
		}

		w.WriteHeader(400)
		w.Write(j)
		return
	}

	_, err = client.AddTracksToPlaylist(groupPlaylistId, spotify.ID(songID))

	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return
	}
	
	track, err := client.GetTrack(spotify.ID(songID))

	if err != nil {
		log.Println(err)
	}

	msg := map[string]interface{} { "type": "addition", "track": track }
	j, err := json.Marshal(msg)
	
	if err != nil {
		log.Println(err)
	}

	WsHub.Broadcast(j, roomCode)
	w.WriteHeader(200)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.FormValue("room-code")

	// See if the room code exists
	found := DoesRoomExist(Db, roomCode)

	if !found {
		log.Printf("Room code %s not found.", roomCode)
		http.Redirect(w, r, "/static/room-not-found.html", http.StatusSeeOther)
		return
	}

	// Save the room code to this users session
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	session.Values["roomCode"] = roomCode
	session.Save(r, w)

	// Redirect to that room	
	http.Redirect(w, r, "/room/" + roomCode, http.StatusSeeOther)
}

func CreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {	
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	client := auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	description := "Playlist for Spotify Group Queue written by Aaron Raff."

	_, err = client.CreatePlaylistForUser(user.ID, "GroupQueue", description, true)

	if err != nil {
		w.WriteHeader(400)
	} else {
		// Success
		w.WriteHeader(200)
	}
}

func VetoHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	roomCode := r.FormValue("roomCode")

	id, ok := session.Values["id"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	UStore.CastUserVote(id, roomCode)
	voteCount := strconv.Itoa(UStore.GetVoteCount(roomCode))

	// Update the front end
	msg := map[string]string { "type": "vetoCountUpdate", "count": voteCount }
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	WsHub.Broadcast(j, roomCode)

	w.WriteHeader(200)
}
