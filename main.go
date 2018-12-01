package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"encoding/gob"
	"github.com/gorilla/sessions"
	"github.com/gorilla/context"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var baseUrl string
var key = []byte("test-key")
var store = sessions.NewCookieStore(key)

type ProfileData struct {
	Username string
}

var redirectURI = "http://localhost:8080/spotify-callback"
var auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadEmail)
var ch = make(chan* spotify.Client)
var state = "testState"

// https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/auth.go
func init() {
	gob.Register(&oauth2.Token{})	
}

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/login", loginHandler);
	http.HandleFunc("/spotify-callback", spotifyCallbackHandler)
	http.HandleFunc("/profile", profileHandler)
	http.ListenAndServe(":" + port, context.ClearHandler(http.DefaultServeMux))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Please log into <a href=%s target=\"_blank\">Spotify</a>", url)	
}

func spotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	
	st := r.FormValue("state")
	if st != state {
		http.NotFound(w, r)
		log.Fatal("State mismatch.")
	}


	session, _ := store.Get(r, "groupQueue")
	session.Values["token"] = tok
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	client := auth.NewClient(tok)
	user, _ := client.CurrentUser()
	fmt.Fprintf(w, "Hello %s!",  user.ID)
}
