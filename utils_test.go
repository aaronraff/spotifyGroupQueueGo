package main

import (
	"fmt"
	"testing"
	"github.com/zmb3/spotify"
)

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
