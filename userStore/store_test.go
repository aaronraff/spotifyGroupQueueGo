package userStore

import (
	"fmt"
	"testing"
	"github.com/gorilla/websocket"
)

var store *Store

func init() {
	store = NewStore()
}

func TestAddUser(t *testing.T) {
	testCases := []struct {
					id string
					roomCode string
					conn *websocket.Conn
				}{
					{ "11111", "123ABC", &websocket.Conn{} },
					{ "22222", "456DEF", &websocket.Conn{} },
					{ "33333", "123ABC", &websocket.Conn{} },
					{ "22222", "123ABC", &websocket.Conn{} },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %s, roomCode: %s, expected websocket.Conn: %p", 
						  tc.id, tc.roomCode, tc.conn),
			func(t *testing.T) {	
				store.AddUser(tc.id, tc.roomCode, tc.conn)
				actual := store.UserExists(tc.id, tc.roomCode) 

				if actual != true {
					t.Errorf("got %t, want true", actual)
				}
		})
	}
}

func TestRemoveUser(t *testing.T) {
	testCases := []struct {
					id string
					roomCode string
					conn *websocket.Conn
				}{
					{ "33333", "123ABC", &websocket.Conn{} },
					{ "33333", "456DEF", &websocket.Conn{} },
					{ "44444", "123ABC", &websocket.Conn{} },
					{ "55555", "123ABC", &websocket.Conn{} },
				}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %s, roomCode: %s, expected websocket.Conn: %p", 
						  tc.id, tc.roomCode, tc.conn),
			func(t *testing.T) {	
				store.AddUser(tc.id, tc.roomCode, tc.conn)
				store.RemoveUser(tc.id, tc.roomCode)
				actual := store.UserExists(tc.id, tc.roomCode) 

				if actual != false {
					t.Errorf("got %t, want false", actual)
				}
		})
	}
}

func TestCastUserVote(t *testing.T) {
	testCases := []struct {
					id string
					roomCode string
					voteCount int
				}{
					{ "66666", "123ABC", 1 },
					{ "77777", "123ABC", 2 },
					{ "88888", "123ABC", 3 },
					{ "66666", "456DEF", 1 },
				}

	// First just add the users		
	for _, tc := range testCases {	
		store.AddUser(tc.id, tc.roomCode, nil)
	}

	// Then cast votes and test
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %s, roomCode: %s, expected vote count: %d", 
						  tc.id, tc.roomCode, tc.voteCount),
			func(t *testing.T) {	
				store.CastUserVote(tc.id, tc.roomCode)
				actual := store.GetVoteCount(tc.roomCode) 

				if actual != tc.voteCount {
					t.Errorf("got %d, want %d", actual, tc.voteCount)
				}
		})
	}
}

func TestShouldSkip(t *testing.T) {
	testCases := []struct {
					id string
					roomCode string
					shouldSkip bool
				}{
					{ "99999", "678ABD", false },
					{ "10101", "678ABD", false },
					{ "12121", "678ABD", true },
					{ "13131", "678ABD", true },
					{ "66666", "101LAC", true },
				}

	// First just add the users		
	for _, tc := range testCases {	
		store.AddUser(tc.id, tc.roomCode, nil)
	}

	// Then cast votes and test
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("id: %s, roomCode: %s, expected should skip: %t", 
						  tc.id, tc.roomCode, tc.shouldSkip),
			func(t *testing.T) {	
				store.CastUserVote(tc.id, tc.roomCode)
				actual := store.ShouldSkip(tc.roomCode) 

				if actual != tc.shouldSkip {
					t.Errorf("got %t, want %t", actual, tc.shouldSkip)
				}
		})
	}
}

func TestResetUsersVote(t *testing.T) {
	testCases := []struct {
					id string
					roomCode string
				}{
					{ "14141", "109AAA" },
					{ "15151", "109AAA" },
					{ "16161", "109AAA" },
					{ "17171", "109AAA" },
				}

	// First just add the users	and vote
	for _, tc := range testCases {	
		store.AddUser(tc.id, tc.roomCode, nil)
		store.CastUserVote(tc.id, tc.roomCode)
	}

	// Then reset user votes
	t.Run(fmt.Sprintf("After skipping, expected vote count: 0"),
		func(t *testing.T) {
			store.ResetUsersVote("109AAA")
			actual := store.GetVoteCount("109AAA") 

			if actual != 0 {
				t.Errorf("got %d, want 0", actual)
			}
	})
}

func TestUserHasVoted(t *testing.T) {
	id1 := "21212"
	id2 := "31313"
	roomCode := "111BBB"
	
	store.AddUser(id1, roomCode, nil)
	store.AddUser(id2, roomCode, nil)

	store.CastUserVote(id1, roomCode)

	t.Run(fmt.Sprintf("id: %s, expected has voted: true", id1),
		func(t *testing.T) {
			actual := store.UserHasVoted(id1, roomCode)

			if actual != true {
				t.Errorf("got %t, want true", actual)
			}
	})

	t.Run(fmt.Sprintf("id: %s, expected has voted: false", id2),
		func(t *testing.T) {
			actual := store.UserHasVoted(id2, roomCode)

			if actual != false {
				t.Errorf("got %t, want true", actual)
			}
	})
}
