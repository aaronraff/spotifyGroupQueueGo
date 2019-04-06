package sessionStore

import (	
	"golang.org/x/oauth2"
)

var store map[string]*oauth2.Token

func init() {
	store = make(map[string]*oauth2.Token)
}

func RegisterUser(uuid string, tok *oauth2.Token) {
	store[uuid] = tok
}

func DestroyUser(uuid string) {
	delete(store, uuid)
}

func GetToken(uuid string) *oauth2.Token {
	return store[uuid]
}
