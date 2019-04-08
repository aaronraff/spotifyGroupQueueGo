package userStore

import (
	"log"
	"github.com/gorilla/websocket"
)

type Store struct {
	users map[string]map[string]*UserInfo
	voteCount map[string]int
	notifySkip map[string]chan bool 
}

type UserInfo struct {
	hasVoted bool
	conn *websocket.Conn
}

// NewStore creates a Store struct and returns a pointer to that struct.
func NewStore() *Store {	
	return &Store { 
		users: make(map[string]map[string]*UserInfo), 
		voteCount: make(map[string]int),
		notifySkip: make(map[string]chan bool),
	}
}

// AddChannel adds a channel (used to notify the poller that a track should
// be skipped) to the notifySkip map of the Store for the specified roomCode.
func (s *Store) AddChannel(roomCode string) chan bool {
	s.notifySkip[roomCode] = make(chan bool)

	// Return the channel to be used by the song poller
	return s.notifySkip[roomCode]
}

// UserExists checks if a user (specified by id) exists in the Store's users
// map. The map maps roomCodes to maps of user ids to UserInfo structs.
func (s *Store) UserExists(id string, roomCode string) bool {
	if _, ok := s.users[roomCode][id]; ok {
		return true
	}

	return false
}

// AddUser adds a user id to the Store's users map. It also stores their
// web socket connection in a UserInfo struct (which is mapped to the user's
// id).
func (s *Store) AddUser(id string, roomCode string, wsConn *websocket.Conn) {
	if _, ok := s.users[roomCode]; !ok {
		// Need to initialize a new map if it has not been created
		s.users[roomCode] = make(map[string]*UserInfo)
	}
	
	s.users[roomCode][id] = &UserInfo{ hasVoted: false, conn: wsConn }
}

// UpdateUserConn is used to update the websocket connection for a specific
// user (and room).
func (s *Store) UpdateUserConn(id string, roomCode string, wsConn *websocket.Conn) {
	s.users[roomCode][id].conn = wsConn
}

// RemoveUser removes a user (specified by id) from the Store's users map.
func (s *Store) RemoveUser(id string, roomCode string) {
	delete(s.users[roomCode], id)
}

// CastUserVote updates the user's UserInfo struct's hasVoted property to true.
func (s *Store) CastUserVote(id string, roomCode string) {
	prevVal := s.users[roomCode][id].hasVoted
	s.users[roomCode][id].hasVoted = true

	// Only update count if they haven't voted yet
	if prevVal == false {
		log.Printf("Vote casted by %s", id)
		s.voteCount[roomCode]++
	}
}

// ShouldSkip returns whether or not the current song for the specified 
// roomCode should be skipped. This is based on if the number of votes in the
// room is greater than half of the room's user count.
func (s *Store) ShouldSkip(roomCode string) bool {	
	if(s.GetVoteCount(roomCode) > (s.GetTotalUserCount(roomCode)/2)) {
		return true
	}

	return false;
}

// NotifySkip sends true to the specified roomCode's shouldSkip channel. This
// channel is being listened to by the Poller.
func (s *Store) NotifySkip(roomCode string) {
	s.notifySkip[roomCode] <- true
}

// UserHasVoted returns whether or not the specified user (based on id) has
// voted in the room that was specified by the roomCode.
func (s *Store) UserHasVoted(id string, roomCode string) bool {
	if val, ok := s.users[roomCode][id]; ok {
		return val.hasVoted
	}

	return false
}

// ResetUsersVote resets all of the user's votes for the specified roomCode.
func (s *Store) ResetUsersVote(roomCode string) {
	log.Printf("Reseting user votes for room: %s", roomCode)
	for id := range s.users[roomCode] {
		s.users[roomCode][id].hasVoted = false	
	}

	s.voteCount[roomCode] = 0
}

// GetTotalUserCount returns the number of users in the room specified by
// the roomCode.
func (s *Store) GetTotalUserCount(roomCode string) int {
	return len(s.users[roomCode])
}

// GetVoteCount returns the number of votes in the room specified by
// the roomCode.
func (s *Store) GetVoteCount(roomCode string) int {
	return s.voteCount[roomCode]
}

// IsActiveConn returns whether or not the specified websocket connection is
// active in the room specified by roomCode. Basically in this case, active
// means it is a user's most recently connected websocket.
func (s *Store) IsActiveConn(roomCode string, wsConn *websocket.Conn) bool {
	for _, info := range s.users[roomCode] {
		if info.conn == wsConn {
			return true
		}
	}

	return false
}

// RemoveRoom removes the room (specified by roomCode) from the Store.
func (s *Store) RemoveRoom(roomCode string) {
	delete(s.notifySkip, roomCode)	
	delete(s.users, roomCode)	
}
