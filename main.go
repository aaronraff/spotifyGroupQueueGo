package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == nil {
		port = 8080
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":" + port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Testing");
}


