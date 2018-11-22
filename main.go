package main

import (
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
)

func main() {
	// For Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":" + port, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	res, _ := ioutil.ReadFile("test.html")
	fmt.Fprintf(w, string(res));
}


