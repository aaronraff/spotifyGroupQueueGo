package main

import (
	"log"
	"net/http"
	"encoding/json"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

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

	user, _ := client.CurrentUser()
	playlists, err := client.GetPlaylistsForUser(user.ID)

	if err != nil {
		log.Fatal(err)
	}

	var groupPlaylist *spotify.SimplePlaylist

	for _, playlist := range playlists.Playlists {
		if playlist.Name == "GroupQueue" {
			groupPlaylist = &playlist			
			break;
		}
	}

	_, err = client.AddTracksToPlaylist(groupPlaylist.ID, spotify.ID(songID))

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
		return
	}

	// Save the room code to this users session
	session, _ := Store.Get(r, "groupQueue")
	session.Values["roomCode"] = roomCode
	session.Save(r, w)

	// Redirect to that room	
	http.Redirect(w, r, "/room/" + roomCode, http.StatusSeeOther)
}
