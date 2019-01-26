package main

import (
	"fmt"
	"testing"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func TestGetTokenFromCode(t *testing.T) {
	Rooms["user1"] = RoomInfo { code: "abc", tok: &oauth2.Token { AccessToken: "token1" }}
	Rooms["user2"] = RoomInfo { code: "bcd", tok: &oauth2.Token { AccessToken: "token2" }}
	Rooms["user3"] = RoomInfo { code: "cde", tok: &oauth2.Token { AccessToken: "token3" }}

	testCases := []struct {
					code string
					expected *oauth2.Token
				}{
					{	"bcd", Rooms["user2"].tok },
					{	"zfg", nil },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("code: %s, expected: %v", tc.code, tc.expected),
			func(t *testing.T) {	
				actual := GetTokenFromCode(tc.code)

				if actual != tc.expected {
					t.Errorf("got %v, want %v", actual, tc.expected)
				}
		})
	}
}

func TestIsSongPresent(t *testing.T) {
	tracks := []spotify.PlaylistTrack {
				spotify.PlaylistTrack { 
					Track: spotify.FullTrack { SimpleTrack: spotify.SimpleTrack { ID: spotify.ID(1) } },
				},
				spotify.PlaylistTrack { 
					Track: spotify.FullTrack { SimpleTrack: spotify.SimpleTrack { ID: spotify.ID(2) } },
				},
				spotify.PlaylistTrack { 
					Track: spotify.FullTrack { SimpleTrack: spotify.SimpleTrack { ID: spotify.ID(3) } },
				},
			}

	testCases := []struct {
					songID spotify.ID
					expected bool
				}{
					{	spotify.ID(1), true },
					{	spotify.ID(4), false },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("songID: %s, expected: %t", string(tc.songID), tc.expected),
			func(t *testing.T) {	
				actual := IsSongPresent(tracks, string(tc.songID))

				if actual != tc.expected {
					t.Errorf("got %t, want %t", actual, tc.expected)
				}
		})
	}
}
