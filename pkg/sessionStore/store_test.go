package sessionStore

import (
	"testing"
	"fmt"
	"golang.org/x/oauth2"
)

func TestRegisterUser(t *testing.T) {
	testCases := []struct {
					uuid string
					token *oauth2.Token
				}{
					{	"11111", &oauth2.Token{} },
					{	"22222", &oauth2.Token{} },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("UUID: %s, expected Token: %p", tc.uuid, tc.token),
			func(t *testing.T) {	
				RegisterUser(tc.uuid, tc.token)
				actual := GetToken(tc.uuid)
				if actual != tc.token {
					t.Errorf("got %p, want %p", actual, tc.token)
				}
		})
	}
}

func TestDestroyUser(t *testing.T) {
	testCases := []struct {
					uuid string
					token *oauth2.Token
				}{
					{	"33333", nil },
					{	"44444", nil },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("UUID: %s, expected Token: %p", tc.uuid, tc.token),
			func(t *testing.T) {	
				RegisterUser(tc.uuid, tc.token)
				DestroyUser(tc.uuid)
				actual := GetToken(tc.uuid)
				if actual != tc.token {
					t.Errorf("got %p, want %p", actual, tc.token)
				}
		})
	}
}
