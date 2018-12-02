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

	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

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

	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

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
