package main

import (	
	"golang.org/x/oauth2"
)

func GetTokenFromCode(roomCode string) *oauth2.Token {
	for _, v := range Rooms	{
		if v.code == roomCode {
			tok := v.tok
			return tok
		}
	}

	return nil
}
