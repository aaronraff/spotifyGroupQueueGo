package router

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/aaronraff/spotifyGroupQueueGo/internal/pkg/utils"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/sessionStore"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/userStore"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/workerStore"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/wsHub"
	"github.com/gorilla/sessions"
	"github.com/zmb3/spotify"
)

type Router struct {
	db        *sql.DB
	hub       *wsHub.Hub
	store     sessions.Store
	auth      *spotify.Authenticator
	userStore *userStore.Store
}

type pageInfo struct {
	User           interface{}
	Code           string
	IsActive       bool
	IsOwner        bool
	QueueSongs     []spotify.PlaylistTrack
	PlaylistExists bool
	IsLoggedIn     bool
	HasVetoed      bool
	VetoCount      int
	UserCount      int
	ShareableLink  string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func NewRouter(db *sql.DB, hub *wsHub.Hub, store sessions.Store, auth *spotify.Authenticator,
	userStore *userStore.Store) *Router {
	return &Router{
		db,
		hub,
		store,
		auth,
		userStore,
	}
}

func (router *Router) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func (router *Router) LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	// There is a user logged in already
	if tok != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	} else {
		// CSRF Protection with state
		b := make([]byte, 20)
		rand.Read(b)
		state := base64.StdEncoding.EncodeToString(b)

		session.Values["state"] = state
		session.Save(r, w)

		url := router.auth.AuthURL(state)
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, url)
	}
}

func (router *Router) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	client := router.auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	// Remove the room from the map
	utils.DeleteRoom(router.db, string(user.ID))

	// Invalidate session
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (router *Router) SpotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	state, ok := session.Values["state"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok, err := router.auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}

	// CSRF Protection with state
	st := r.FormValue("state")
	if st != state {
		http.NotFound(w, r)
		log.Fatal("State mismatch.")
	}

	uuid := utils.GenerateUUID()

	// Register the user in the session store
	sessionStore.RegisterUser(uuid, tok)

	// Store the UUID
	session.Values["uuid"] = uuid
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (router *Router) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/profile.html"))

	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	// Generate an id for the session if one does not exist
	id, ok := session.Values["id"].(string)

	if !ok {
		// Need a UUID for each user
		uuid := utils.GenerateUUID()

		session.Values["id"] = uuid
		id = uuid

		session.Save(r, w)
	}

	isLoggedIn := false

	tok := sessionStore.GetToken(uuid)

	if tok != nil {
		isLoggedIn = true
	} else {
		// Need to login to see the profile page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	client := router.auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	roomCode := utils.GetRoomCode(router.db, string(user.ID))
	active := utils.DoesRoomExist(router.db, roomCode)

	queueSongs, playlistExists := utils.GetQueueSongs(&client)

	hasVetoed := router.userStore.UserHasVoted(id, roomCode)

	pInfo := pageInfo{
		User:           user,
		Code:           roomCode,
		IsActive:       active,
		IsOwner:        true,
		QueueSongs:     queueSongs.Tracks,
		PlaylistExists: playlistExists,
		IsLoggedIn:     isLoggedIn,
		HasVetoed:      hasVetoed,
		VetoCount:      router.userStore.GetVoteCount(roomCode),
		UserCount:      router.userStore.GetTotalUserCount(roomCode),
		ShareableLink:  utils.GenerateShareableLink(r, roomCode),
	}

	if err = tmpl.Execute(w, pInfo); err != nil {
		log.Println(err)
	}
}

func (router *Router) RoomHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	// Generate an id for the session if one does not exist
	id, ok := session.Values["id"].(string)

	if !ok {
		// Need a UUID for each user
		uuid := utils.GenerateUUID()

		// Need to cut off at 7 chars (base64 can be longer)
		session.Values["id"] = uuid
		id = uuid

		session.Save(r, w)
	}

	roomCode := r.URL.Path[len("/room/"):]

	// See if the room code exists
	found := utils.DoesRoomExist(router.db, roomCode)

	if !found {
		log.Printf("Room code %s not found.", roomCode)
		http.Redirect(w, r, "/static/room-not-found.html", http.StatusSeeOther)
		return
	}

	// Get the token
	tok := utils.GetTokenFromCode(router.db, roomCode)

	// Need a client to get the songs in the playlist
	client := router.auth.NewClient(tok)

	groupPlaylistID := utils.GetPlaylistIdByName(&client, "GroupQueue")
	queueSongs, err := client.GetPlaylistTracks(groupPlaylistID)

	if err != nil {
		log.Println(err)
	}

	hasVetoed := router.userStore.UserHasVoted(id, roomCode)

	pInfo := pageInfo{
		User:           struct{ ID string }{string(roomCode)},
		Code:           string(roomCode),
		IsActive:       true,
		IsOwner:        false,
		QueueSongs:     queueSongs.Tracks,
		PlaylistExists: false,
		IsLoggedIn:     false,
		HasVetoed:      hasVetoed,
		VetoCount:      router.userStore.GetVoteCount(roomCode),
		UserCount:      router.userStore.GetTotalUserCount(roomCode),
		ShareableLink:  utils.GenerateShareableLink(r, roomCode),
	}

	tmpl := template.Must(template.ParseFiles("templates/profile.html"))
	tmpl.Execute(w, pInfo)
}

func (router *Router) OpenRoomHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	client := router.auth.NewClient(tok)
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
	utils.InsertRoom(router.db, roomCode, string(user.ID), tok)

	// Make sure there is atleast one song in the playlist
	utils.CheckAddRandomSong(&client, roomCode, router.hub)

	// Update front end with room code for /room/start
	msg := map[string]string{"roomCode": roomCode}
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
func (router *Router) StartPollerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	roomCode := r.FormValue("roomCode")

	client := router.auth.NewClient(tok)

	utils.StartFirstTrack(&client)

	notifyChan := router.userStore.AddChannel(roomCode)
	cancelChan := workerStore.AddPoller(roomCode)
	go utils.PollPlayerForRemoval(&client, roomCode, router.hub, router.userStore, notifyChan, cancelChan)

	w.WriteHeader(200)
}

func (router *Router) CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	roomCode := r.FormValue("roomCode")

	client := router.auth.NewClient(tok)
	user, err := client.CurrentUser()

	if err != nil {
		log.Println(err)
	}

	// Stop the poller
	cancelChan := workerStore.GetPollerChan(roomCode)
	cancelChan <- true
	workerStore.RemovePoller(roomCode)

	// Remove the room from the DB
	utils.DeleteRoom(router.db, string(user.ID))
	router.userStore.RemoveRoom(roomCode)

	msg := map[string]interface{}{"type": "roomClosed"}
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	router.hub.Broadcast(j, roomCode)

	// Success
	w.WriteHeader(200)
}

func (router *Router) SearchHandler(w http.ResponseWriter, r *http.Request) {
	songName := r.FormValue("songName")
	roomCode := r.FormValue("roomCode")

	// Get the token
	tok := utils.GetTokenFromCode(router.db, roomCode)

	client := router.auth.NewClient(tok)

	results, err := client.Search(songName, spotify.SearchTypeTrack)

	if err != nil {
		log.Println(err)
	}

	resJSON, err := json.Marshal(results.Tracks.Tracks)

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")

	// Send response to client
	w.Write(resJSON)
}

func (router *Router) AddToQueueHandler(w http.ResponseWriter, r *http.Request) {
	songID := r.FormValue("songID")
	roomCode := r.FormValue("roomCode")

	// Get the token
	tok := utils.GetTokenFromCode(router.db, roomCode)

	client := router.auth.NewClient(tok)
	groupPlaylistID := utils.GetPlaylistIdByName(&client, "GroupQueue")

	tracks, err := client.GetPlaylistTracks(groupPlaylistID)

	if err != nil {
		log.Println(err)
	}

	if utils.IsSongPresent(tracks.Tracks, songID) {
		msg := map[string]interface{}{"msg": "This song is already in the queue."}
		j, err := json.Marshal(msg)

		if err != nil {
			log.Println(err)
		}

		w.WriteHeader(400)
		w.Write(j)
		return
	}

	_, err = client.AddTracksToPlaylist(groupPlaylistID, spotify.ID(songID))

	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return
	}

	track, err := client.GetTrack(spotify.ID(songID))

	if err != nil {
		log.Println(err)
	}

	msg := map[string]interface{}{"type": "addition", "track": track}
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	router.hub.Broadcast(j, roomCode)
	w.WriteHeader(200)
}

func (router *Router) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.FormValue("room-code")

	// See if the room code exists
	found := utils.DoesRoomExist(router.db, roomCode)

	if !found {
		log.Printf("Room code %s not found.", roomCode)
		http.Redirect(w, r, "/static/room-not-found.html", http.StatusSeeOther)
		return
	}

	// Save the room code to this users session
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	session.Values["roomCode"] = roomCode
	session.Save(r, w)

	// Redirect to that room
	http.Redirect(w, r, "/room/"+roomCode, http.StatusSeeOther)
}

func (router *Router) CreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	uuid, ok := session.Values["uuid"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	tok := sessionStore.GetToken(uuid)

	client := router.auth.NewClient(tok)
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

func (router *Router) VetoHandler(w http.ResponseWriter, r *http.Request) {
	session, err := router.store.Get(r, "groupQueue")

	if err != nil {
		log.Println(err)
	}

	roomCode := r.FormValue("roomCode")

	id, ok := session.Values["id"].(string)

	if !ok {
		log.Println("Session value is not of type string")
	}

	router.userStore.CastUserVote(id, roomCode)
	if router.userStore.ShouldSkip(roomCode) {
		router.userStore.ResetUsersVote(roomCode)
		// Let the poller know we are skipping a song
		router.userStore.NotifySkip(roomCode)
	}

	voteCount := strconv.Itoa(router.userStore.GetVoteCount(roomCode))

	// Update the front end
	msg := map[string]string{"type": "vetoCountUpdate", "count": voteCount}
	j, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
	}

	router.hub.Broadcast(j, roomCode)

	w.WriteHeader(200)
}
