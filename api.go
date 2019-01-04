package main

import (
	"log"
	"net/http"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func OpenRoomHandler(w http.ResponseWriter, r *http.Request) {
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
	val := RoomInfo{str[:7], tok}
	Rooms[user.ID] = val

	go PollPlayerForRemoval(&client)

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
		log.Fatal(err)
	}

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
		http.Redirect(w, r, "/room-not-found.html", http.StatusSeeOther)
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
