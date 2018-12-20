package main

import (
	"log"
	"os"
	"net/http"
	"encoding/gob"
	"crypto/rand"
	"text/template"
	"encoding/base64"
	"github.com/gorilla/sessions"
	"github.com/gorilla/context"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type RoomInfo struct {
	code string
	tok* oauth2.Token
}

var key = []byte("test-key")

var Rooms = make(map[string]RoomInfo)

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
	http.HandleFunc("/add", AddToQueueHandler)
	http.HandleFunc("/join", JoinRoomHandler)
	http.HandleFunc("/room/", roomHandler)
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
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, url)
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

	val, ok := Rooms[user.ID]
	
	if !ok {
		// No room code exists for this user
		// Generate one
		code := make([]byte, 7)
		rand.Read(code)

		// Need to make it base64
		str := base64.StdEncoding.EncodeToString(code)

		// Need to cut off at 7 chars (base64 can be longer)
		val = RoomInfo{str[:7], tok}
		Rooms[user.ID] = val
	}

	tmpl.Execute(w, map[string]interface{} {"user": user, "code": val.code})
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Path[len("/room/"):]
	
	// See if the room code exists
	found := false
	for _, v := range Rooms {
		if v.code == roomCode {
			found = true
			break
		}
	}

	if !found {	
		log.Printf("Room code %s not found.", roomCode)
		return
	}

	log.Print("testing")
	tmpl := template.Must(template.ParseFiles("profile.html"))
	tmpl.Execute(w, map[string]interface{} {"user": struct{ID string} {"test"}, "code": string(roomCode)})
}
