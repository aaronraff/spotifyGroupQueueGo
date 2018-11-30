package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"encoding/json"
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
	fmt.Println(url)
	fmt.Println("Please log into Spotify", url)

	// Wait for auth to complete
	client := <-ch

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, "You are logged in as: ", user.ID)
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

	client := auth.NewClient(tok)
	ch <- &client

	tokJson, _ := json.Marshal(tok)

	session, _ := store.Get(r, "groupQueue")
	session.Values["token"] = tokJson
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "groupQueue")
	var tok oauth2.Token
	val, _ := session.Values["token"].([]byte)
	json.Unmarshal(val, tok) 

	client := auth.NewClient(&tok)
	user, _ := client.CurrentUser()
	fmt.Fprintf(w, "Hello %s!",  user.ID)
}
