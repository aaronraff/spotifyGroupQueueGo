package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"encoding/json"
	"bytes"
)

var baseUrl string

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/login", loginHandler);
	http.HandleFunc("/spotify-callback", spotifyCallbackHandler)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":" + port, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	clientId := os.Getenv("CLIENTID")
	baseUrl = "http://" + r.Host
	redirectUri := url.QueryEscape(baseUrl + "/spotify-callback")
	scopes := url.QueryEscape("user-modify-playback-state user-read-playback-state")
	url := "https://accounts.spotify.com/authorize"	+
		   "?response_type=code" + "&client_id=" + clientId +
		   "&scope=" + scopes + "&redirect_uri=" + redirectUri
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func spotifyCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Used for obtaining an access token
	fmt.Printf("Made it.")
	code := r.URL.Query().Get("code")
	body := make(map[string]string)
	body["grant-type"] = "authorization-code"
	body["code"] = code
	bodyJson, _ := json.Marshal(body)
	// Must match URI that was used when requesting code (not actually used)
	body["redirect_uri"] = baseUrl + "/spotify-callback"
	http.Post("https://accounts.spotify.com/api/token", "application/x-www-form-urlencoded", bytes.NewBuffer(bodyJson))
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:] + ".html")
}


