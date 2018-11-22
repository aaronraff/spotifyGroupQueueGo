package main

import (
	"net/http"
	"net/url"
	"os"
)

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/login", loginHandler);
	http.HandleFunc("/", handler)
	http.ListenAndServe(":" + port, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	clientId := os.Getenv("CLIENTID")
	redirectUri:= url.QueryEscape("http://localhost:8080/")
	scopes := url.QueryEscape("user-modify-playback-state user-read-playback-state")
	url := "https://accounts.spotify.com/authorize"	+
		   "?response_type=code" + "&client_id=" + clientId +
		   "&scope=" + scopes + "&redirect_uri=" + redirectUri
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}


