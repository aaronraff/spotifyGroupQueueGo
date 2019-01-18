package main

import (	
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify"
	"log"
	"encoding/json"
	"time"
	"spotifyGroupQueueGo/wsHub"
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

func PollPlayerForRemoval(client *spotify.Client, roomCode string, hub *wsHub.Hub) {
	// Need this to remove tracks
	playlistID := GetPlaylistIdByName(client, "GroupQueue")

	// Used to eventually end the Go routine
	retryCount := 0

	lastPlaying, _ := client.PlayerCurrentlyPlaying()
	for {
		currPlaying, err := client.PlayerCurrentlyPlaying()

		// Nothing is currently being played
		if currPlaying == nil {
			// Wait for something to be playing
			time.Sleep(2 * time.Minute)
			continue
		}

		if err != nil {
			log.Println(err)
			continue
		}

		if currPlaying.Item.ID != lastPlaying.Item.ID {
			// Reset the retry count (we did something)
			retryCount = 0
			client.RemoveTracksFromPlaylist(playlistID, lastPlaying.Item.ID)
			msg := map[string]string { "type": "removal", "trackID": string(lastPlaying.Item.ID) }
			j, err := json.Marshal(msg)
			
			if err != nil {
				log.Fatal(err)
			}

			log.Println(msg)
			hub.Broadcast(j, roomCode)
		}

		// Add 1 sec as a buffer
		timeLeft := currPlaying.Item.Duration - currPlaying.Progress + 1000
		log.Println(timeLeft)
		time.Sleep(time.Duration(timeLeft) * time.Millisecond)
		lastPlaying = currPlaying
		retryCount++

		// Stop trying to remove tracks
		if retryCount > 5 {
			log.Println("Retry count reached, done polling.")
			return
		}
	}
}
