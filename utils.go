package main

import (	
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify"
	"log"
	"time"
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

func GetPlaylistIdByName(client *spotify.Client, playlistName string) spotify.ID {
	user, _ := client.CurrentUser()
	playlists, err := client.GetPlaylistsForUser(user.ID)

	if err != nil {
		log.Fatal(err)
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name == playlistName {
			return playlist.ID
		}
	}

	return ""
}

func PollPlayerForRemoval(client *spotify.Client) {
	// Need this to remove tracks
	playlistID := GetPlaylistIdByName(client, "GroupQueue")

	// Used to eventually end the Go routine
	retryCount := 0

	lastPlaying, _ := client.PlayerCurrentlyPlaying()
	for {
		currPlaying, err := client.PlayerCurrentlyPlaying()

		if err != nil {
			log.Println(err)
			continue
		}

		if currPlaying.Item.ID != lastPlaying.Item.ID {
			log.Println("removing")

			// Reset the retry count (we did something)
			retryCount = 0
			client.RemoveTracksFromPlaylist(playlistID, lastPlaying.Item.ID)
		}

		timeLeft := currPlaying.Item.Duration - currPlaying.Progress
		log.Println(timeLeft)
		time.Sleep(time.Duration(timeLeft) * time.Millisecond)
		lastPlaying = currPlaying
		retryCount++

		// Stop trying to remove tracks
		if retryCount > 5 {
			return
		}
	}
}
