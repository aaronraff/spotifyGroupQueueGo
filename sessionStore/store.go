package sessionStore

import (	
	"golang.org/x/oauth2"
)

var store map[string]*oauth2.Token

func init() {
	store = make(map[string]*oauth2.Token)
}

// RegisterUser adds a uuid along with that user's Spotify token to the session
// store.
func RegisterUser(uuid string, tok *oauth2.Token) {
	store[uuid] = tok
}

// DestroyUser removes the entry with the given uuid from the session store.
func DestroyUser(uuid string) {
	delete(store, uuid)
}

// GetToken gets the Spotify token associated with the given uuid from the
// session store.
func GetToken(uuid string) *oauth2.Token {
	return store[uuid]
}
