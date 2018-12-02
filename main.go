package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"encoding/gob"
	"text/template"
	"github.com/gorilla/sessions"
	"github.com/gorilla/context"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var aseUrl string
var key = []byte("test-key")

// Uppercase so it can be accessed by the api
var Store = sessions.NewCookieStore(key)

var redirectURI = "http://localhost:8080/spotify-callback"
var auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadEmail, spotify.ScopePlaylistModifyPublic)
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
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/spotify-callback", spotifyCallbackHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/search", SearchHandler)
	http.HandleFunc("/add", AddToQueueHandler )
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":" + port, context.ClearHandler(http.DefaultServeMux))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.AuthURL(state)

	session, _ := Store.Get(r, "groupQueue")
	tok := session.Values["token"]

	// There is a user logged in already
	if tok != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)	
	} else {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "Please log into <a href=%s>Spotify</a>", url)	
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "groupQueue")

	// Invalidate session
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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


	session, _ := Store.Get(r, "groupQueue")
	session.Values["token"] = tok
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("profile.html"))

	session, _ := Store.Get(r, "groupQueue")
	tok, _ := session.Values["token"].(*oauth2.Token)

	client := auth.NewClient(tok)
	user, _ := client.CurrentUser()

	tmpl.Execute(w, user)
}

