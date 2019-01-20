package main

import (
	"log"
	"strconv"
	"net/http"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"spotifyGroupQueueGo/wsHub"
)

func OpenRoomHandler(hub *wsHub.Hub, w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	client := auth.NewClient(tok)
	user, _ := client.CurrentUser()

	// Generate a code for the new room
	code := make([]byte, 7)
	rand.Read(code)

	// Need to make it base64
	str := base64.StdEncoding.EncodeToString(code)

	// Need to cut off at 7 chars (base64 can be longer)
	roomCode := str[:7]

	val := RoomInfo{roomCode, tok}
	Rooms[user.ID] = val

	notifyChan := UStore.AddChannel(roomCode)

	go PollPlayerForRemoval(&client, roomCode, hub, notifyChan)

	// Success
	w.WriteHeader(200)
}

func CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	client := auth.NewClient(tok)
	user, _ := client.CurrentUser()

	// Remove the room from the map
	delete(Rooms, user.ID)

	// Success
	w.WriteHeader(200)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	songName := r.FormValue("songName")
	roomCode := r.FormValue("roomCode")

	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)
	
	// Must be a guest in someone's room
	if tok == nil {
		// Get the token
		tok = GetTokenFromCode(roomCode)
	}

	client := auth.NewClient(tok)

	results, err := client.Search(songName, spotify.SearchTypeTrack)
	if err != nil {
		log.Fatal(err)
	}

	resJson, _ := json.Marshal(results.Tracks.Tracks)

	w.Header().Set("Content-Type", "application/json")

	// Send response to client
	w.Write(resJson)
}

func AddToQueueHandler(w http.ResponseWriter, r *http.Request) {
	songID := r.FormValue("songID")
	roomCode := r.FormValue("roomCode")

	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	// Must be a guest in someone's room
	if tok == nil {
		// Get the token
		tok = GetTokenFromCode(roomCode)
	}

	client := auth.NewClient(tok)
	groupPlaylistId := GetPlaylistIdByName(&client, "GroupQueue")

	_, err := client.AddTracksToPlaylist(groupPlaylistId, spotify.ID(songID))

	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return
	}
	
	track, err := client.GetTrack(spotify.ID(songID))
	msg := map[string]interface{} { "type": "addition", "track": track }
	j, err := json.Marshal(msg)
	
	if err != nil {
		log.Fatal(err)
	}

	WsHub.Broadcast(j, roomCode)
	w.WriteHeader(200)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.FormValue("room-code")

	// See if the room code exists
	found := false
	for _, v := range Rooms {
		if v.code == roomCode {
			found = true
			break
		}
	}

	if !found {
		log.Printf("Room code %s not found.", roomCode)
		http.Redirect(w, r, "/static/room-not-found.html", http.StatusSeeOther)
		return
	}

	// Save the room code to this users session
	session, _ := Store.Get(r, "groupQueue")
	session.Values["roomCode"] = roomCode
	session.Save(r, w)

	// Redirect to that room	
	http.Redirect(w, r, "/room/" + roomCode, http.StatusSeeOther)
}

func CreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {	
	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	client := auth.NewClient(tok)
	user, _ := client.CurrentUser()

	description := "Playlist for Spotify Group Queue written by Aaron Raff."

	_, err := client.CreatePlaylistForUser(user.ID, "GroupQueue", description, true)

	if err != nil {
		w.WriteHeader(400)
	} else {
		// Success
		w.WriteHeader(200)
	}
}

func VetoHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "groupQueue")
	roomCode := r.FormValue("roomCode")

	id, _ := session.Values["id"].(string)

	UStore.CastUserVote(id, roomCode)
	voteCount := strconv.Itoa(UStore.GetVoteCount(roomCode))

	// Update the front end
	msg := map[string]string { "type": "vetoCountUpdate", "count": voteCount }
	j, _ := json.Marshal(msg)
	log.Println(msg)

	WsHub.Broadcast(j, roomCode)

	w.WriteHeader(200)
}
