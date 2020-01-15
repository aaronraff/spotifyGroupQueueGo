package main

import (
	"database/sql"

	"github.com/aaronraff/spotifyGroupQueueGo/internal/pkg/router"
	"github.com/aaronraff/spotifyGroupQueueGo/internal/pkg/utils"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/userStore"
	"github.com/aaronraff/spotifyGroupQueueGo/pkg/wsHub"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/zmb3/spotify"

	"log"
	"net/http"
	"os"
)

var key = []byte(os.Getenv("SESSION_KEY"))
var store = sessions.NewFilesystemStore("./sessions/", key)

var redirectURI = os.Getenv("redirectURI")
var auth spotify.Authenticator
var myRouter *router.Router

var db *sql.DB
var connString = os.Getenv("DATABASE_URL")

var hub = wsHub.NewHub()
var uStore = userStore.NewStore()

// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/auth.go
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Used for heroku
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/spotify-callback"
	}

	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadEmail, spotify.ScopePlaylistModifyPublic,
		spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadPlaybackState)

	var err error

	// Connect to the DB
	db, err = sql.Open("postgres", connString)

	if err != nil {
		log.Fatal(err)
	}

	myRouter = router.NewRouter(db, hub, store, &auth, uStore)
}

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// If the app restarts on Heroku we need to restart the pollers to
	// avoid disruption
	utils.RestartPollers(db, hub, uStore, &auth)

	http.HandleFunc("/", myRouter.LoginHandler)
	http.HandleFunc("/favicon.ico", myRouter.FaviconHandler)
	http.HandleFunc("/logout", myRouter.LogoutHandler)
	http.HandleFunc("/spotify-callback", myRouter.SpotifyCallbackHandler)

	http.HandleFunc("/profile", myRouter.ProfileHandler)
	http.HandleFunc("/search", myRouter.SearchHandler)
	http.HandleFunc("/add", myRouter.AddToQueueHandler)
	http.HandleFunc("/join", myRouter.JoinRoomHandler)

	http.HandleFunc("/room/open", myRouter.OpenRoomHandler)
	http.HandleFunc("/room/start", myRouter.StartPollerHandler)
	http.HandleFunc("/room/close", myRouter.CloseRoomHandler)
	http.HandleFunc("/room/veto", myRouter.VetoHandler)
	http.HandleFunc("/room/", myRouter.RoomHandler)

	http.HandleFunc("/playlist/create", myRouter.CreatePlaylistHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHub.WsHandler(hub, store, uStore, w, r)
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":"+port, context.ClearHandler(http.DefaultServeMux))
}
