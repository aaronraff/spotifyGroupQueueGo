package main

import (	
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify"
	"log"
)

func GetTokenFromCode(roomCode string) *oauth2.Token {
	for _, v := range Rooms	{
		if v.code == roomCode {
			tok := v.tok
			return tok
		}
	}

	return nil
}

func GetPlaylistIdByName(client spotify.Client, playlistName string) spotify.ID {
	user, _ := client.CurrentUser()
	playlists, err := client.GetPlaylistsForUser(user.ID)

	if err != nil {
		log.Fatal(err)
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name == "GroupQueue" {
			return playlist.ID
		}
	}

	return ""
}
