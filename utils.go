package main

import (	
	"golang.org/x/oauth2"
	"github.com/zmb3/spotify"
	"log"
	"encoding/json"
	"time"
	"math/rand"
	"spotifyGroupQueueGo/wsHub"
)

var topPlaylistId spotify.ID

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
		log.Println(err)
		return ""
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name == playlistName {
			return playlist.ID
		}
	}

	return ""
}

func PollPlayerForRemoval(client *spotify.Client, roomCode string, hub *wsHub.Hub, notifyChan chan bool) {
	// Need this to remove tracks
	playlistID := GetPlaylistIdByName(client, "GroupQueue")
	globalPlaylistID := GetPlaylistIdByName(client, "United States Top 50")
	log.Println(globalPlaylistID)

	rand.Seed(time.Now().UTC().UnixNano())

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

			hub.Broadcast(j, roomCode)

			// Reset vote button on front end
			msg = map[string]string { "type": "resetVote" }
			j, _ = json.Marshal(msg)

			hub.Broadcast(j, roomCode)
		}

		// Check if we need to randomly add a song
		tracks, _ := client.GetPlaylistTracks(playlistID)
		if len(tracks.Tracks) <= 1 {
			if topPlaylistId == "" {
				playlists, err := client.GetCategoryPlaylists("toplists")

				if err != nil {
					log.Println(err)
					continue
				}

				for _, p := range playlists.Playlists {
					if p.Name == "Global Top 50" {
						topPlaylistId = p.ID
						break
					}
				}

				if topPlaylistId == "" {
					log.Println("Could not find Global Top 50 playlist")
					continue
				}
			}

			// Add a random song from top 50
			choices, err := client.GetPlaylistTracks(topPlaylistId)

			if err != nil {
				log.Println(err)
				continue
			}

			selection := choices.Tracks[rand.Intn(50)]

			_, err = client.AddTracksToPlaylist(playlistID, spotify.ID(selection.Track.ID))

			if err != nil {
				log.Println(err)
				continue
			}

			// Update front end
			track, err := client.GetTrack(spotify.ID(selection.Track.ID))
			msg := map[string]interface{} { "type": "addition", "track": track }
			j, err := json.Marshal(msg)
			
			if err != nil {
				log.Fatal(err)
			}

			WsHub.Broadcast(j, roomCode)
		}


		lastPlaying = currPlaying
		
		// Add 1 sec as a buffer
		timeLeft := currPlaying.Item.Duration - currPlaying.Progress + 1000

		select {
			case <-time.After(time.Duration(timeLeft) * time.Millisecond):
				retryCount++
			case <-notifyChan:
				// The song has been vetoed, skip it
				client.Next()
				time.Sleep(1 * time.Second)
				continue
		}

		// Stop trying to remove tracks
		if retryCount > 5 {
			log.Println("Retry count reached, done polling.")
			return
		}
	}
}

func IsSongPresent(tracks []spotify.PlaylistTrack, songId string) bool {
	for _, song := range tracks {
		if song.Track.ID == spotify.ID(songId) {
			return true
		}
	}

	return false
}
