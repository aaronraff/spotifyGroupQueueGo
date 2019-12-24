package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/aaronraff/spotifyGroupQueueGo/userStore"
	"github.com/aaronraff/spotifyGroupQueueGo/workerStore"
	"github.com/aaronraff/spotifyGroupQueueGo/wsHub"
	"github.com/zmb3/spotify"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var topPlaylistId spotify.ID

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UTC().UnixNano())
}

// RestartPollers is used to start up a poller for every currently open room.
// This is mainly used when the application fails and is restarted. That way
// we have a running poller for every room (even after a failure).
func RestartPollers(db *sql.DB, hub *wsHub.Hub, uStore *userStore.Store) {
	codes := GetAllRoomCodes(db)

	for _, roomCode := range codes {
		tok := GetTokenFromCode(db, roomCode)
		client := auth.NewClient(tok)
		notifyChan := UStore.AddChannel(roomCode)
		cancelChan := workerStore.AddPoller(roomCode)
		go PollPlayerForRemoval(&client, roomCode, hub, uStore, notifyChan, cancelChan)
	}
}

// GetPlaylistIdByName return the spotify ID for the specified playlist name.
// This playlist must be one of the user's playlists. If the playlist is not
// found, an empty string is returned.
func GetPlaylistIdByName(client *spotify.Client, playlistName string) spotify.ID {
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
		return ""
	}

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

// PollPlayerForRemoval periodically checks the status of the currently playing
// song to see if it needs to be removed from the playlist. It also skips and
// removes songs from the GroupQueue playlist when it gets notified that the
// current song should be skipped. The cancelChan parameter is used to receive
// notice that the poller should be stopped (ex. the room is closed).
func PollPlayerForRemoval(client *spotify.Client, roomCode string, hub *wsHub.Hub,
	uStore *userStore.Store, notifyChan chan bool,
	cancelChan chan bool) {
	// Need this to remove tracks
	playlistID := GetPlaylistIdByName(client, "GroupQueue")
	retryCount := 0
	lastPlaying, err := client.PlayerCurrentlyPlaying()

	// EOF means nothing is currently playing
	if err != nil && err.Error() != "EOF" {
		log.Println(err)
		return
	}

	for retryCount < 5 {
		currPlaying, err := client.PlayerCurrentlyPlaying()

		// EOF means nothing is currently playing
		if err != nil {
			if err.Error() == "EOF" {
				// Wait for something to be playing
				time.Sleep(15 * time.Second)
			} else {
				log.Println(err)
			}

			retryCount++
			continue
		}

		// Need to also check is anything is playing
		if (currPlaying.Item != nil) && (currPlaying.Item.ID != lastPlaying.Item.ID) {
			// Reset the retry count (we did something)
			retryCount = 0
			client.RemoveTracksFromPlaylist(playlistID, lastPlaying.Item.ID)

			// We are on a new song so we need to reset the votes
			uStore.ResetUsersVote(roomCode)

			// Update the front end to show 0 votes to skip
			// since we just reset the count
			msg := map[string]string{"type": "vetoCountUpdate", "count": "0"}
			j, err := json.Marshal(msg)

			if err != nil {
				log.Println(err)
				continue
			}

			hub.Broadcast(j, roomCode)

			msg = map[string]string{"type": "removal", "trackID": string(lastPlaying.Item.ID)}
			j, err = json.Marshal(msg)

			if err != nil {
				log.Println(err)
				continue
			}

			hub.Broadcast(j, roomCode)

			// Reset vote button on front end
			msg = map[string]string{"type": "resetVote"}
			j, err = json.Marshal(msg)

			if err != nil {
				log.Println(err)
				continue
			}

			hub.Broadcast(j, roomCode)
		}

		// Check if we need to randomly add a song
		checkAddRandomSong(client, roomCode)

		lastPlaying = currPlaying
		// If there is no song currently playing, there should be one starting
		timeLeft := 1000 * 60

		if currPlaying.Item != nil {
			// Add 1 sec as a buffer
			timeLeft = currPlaying.Item.Duration - currPlaying.Progress + 1000
		}

		select {
		case <-time.After(time.Duration(timeLeft) * time.Millisecond):
			retryCount++
		case <-cancelChan:
			// The poller should shut down
			log.Println("Stopping poller")
			return
		case <-notifyChan:
			// Set the last playing
			// Handles edge case where the first song in the queue is skipped
			lastPlaying, err = client.PlayerCurrentlyPlaying()

			// EOF means nothing is currently playing
			if err != nil && err.Error() != "EOF" {
				log.Println(err)
			}

			// The song has been vetoed, skip it
			err = client.Next()

			if err != nil {
				log.Println(err)
			}

			time.Sleep(1 * time.Second)
			continue
		}
	}
}

// checkAddRandom song is used to see if a new song needs to be added to the
// playlist. If there is only one song currently in the playlist (queue), it
// will then add a random song to the playlist.
func checkAddRandomSong(client *spotify.Client, roomCode string) {
	playlistID := GetPlaylistIdByName(client, "GroupQueue")
	tracks, err := client.GetPlaylistTracks(playlistID)

	if err != nil {
		log.Println(err)
	}

	if len(tracks.Tracks) <= 1 {
		if err := addRandomSong(client, playlistID, tracks.Tracks, roomCode); err != nil {
			log.Println(err)
		}
	}
}

// addRandomSong adds a random song to the GroupQueue playlist. It also ensures
// that this song is not already in the playlist (by looking at the tracks
// parameter). The roomCode parameter is used to then broadcast the changes to
// the websocket clients.
func addRandomSong(client *spotify.Client, playlistID spotify.ID, tracks []spotify.PlaylistTrack, roomCode string) error {
	if topPlaylistId == "" {
		playlists, err := client.GetCategoryPlaylists("toplists")

		if err != nil {
			return err
		}

		for _, p := range playlists.Playlists {
			if p.Name == "Global Top 50" {
				topPlaylistId = p.ID
				break
			}
		}

		if topPlaylistId == "" {
			return errors.New("Could not find Global Top 50 playlist")
		}
	}

	// Add a random song from top 50
	choices, err := client.GetPlaylistTracks(topPlaylistId)

	if err != nil {
		return err
	}

	selection := choices.Tracks[rand.Intn(50)]

	// Make sure we don't add a song already in the queue
	for IsSongPresent(tracks, string(selection.Track.ID)) == true {
		selection = choices.Tracks[rand.Intn(50)]
	}

	_, err = client.AddTracksToPlaylist(playlistID, spotify.ID(selection.Track.ID))

	if err != nil {
		return err
	}

	// Update front end
	track, err := client.GetTrack(spotify.ID(selection.Track.ID))
	msg := map[string]interface{}{"type": "addition", "track": track}
	j, err := json.Marshal(msg)

	if err != nil {
		log.Fatal(err)
	}

	WsHub.Broadcast(j, roomCode)

	return nil
}

// IsSongPresent returns whether or not a songId is in an array of Spotify
// tracks.
func IsSongPresent(tracks []spotify.PlaylistTrack, songId string) bool {
	for _, song := range tracks {
		if song.Track.ID == spotify.ID(songId) {
			return true
		}
	}

	return false
}

// generateUUID generates and returns a Universally Unique Identifier.
func generateUUID() string {
	identifier := make([]byte, 7)
	rand.Read(identifier)

	// Need to make it base64
	return base64.StdEncoding.EncodeToString(identifier)
}

// getQueueSongs is used to retrieve the songs currently in the GroupQueue
// playlist. The client param is used to make authenticated requested to the
// Spotify API.
func getQueueSongs(client *spotify.Client) (*spotify.PlaylistTrackPage, bool) {
	groupPlaylistId := GetPlaylistIdByName(client, "GroupQueue")
	playlistExists := true
	queueSongs := new(spotify.PlaylistTrackPage)
	var err error

	// No playlist exists with that name
	if groupPlaylistId == "" {
		playlistExists = false
		queueSongs.Tracks = make([]spotify.PlaylistTrack, 0)
	} else {
		queueSongs, err = client.GetPlaylistTracks(groupPlaylistId)

		if err != nil {
			log.Println(err)
		}
	}

	return queueSongs, playlistExists
}

// generateShareableLink will generate a string (link) to the room using the
// roomCode provided. It uses the passed in Request struct to determine the
// scheme and host information for the link.
func generateShareableLink(r *http.Request, roomCode string) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	return scheme + "://" + r.Host + "/room/" + roomCode
}
