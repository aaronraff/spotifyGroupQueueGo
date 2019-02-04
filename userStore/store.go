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

func NewStore() *Store {	
	return &Store { 
		users: make(map[string]map[string]*UserInfo), 
		voteCount: make(map[string]int),
		notifySkip: make(map[string]chan bool),
	}
}

func (s *Store) AddChannel(roomCode string) chan bool {
	s.notifySkip[roomCode] = make(chan bool)

	// Return the channel to be used by the song poller
	return s.notifySkip[roomCode]
}

func (s *Store) UserExists(id string, roomCode string) bool {
	if _, ok := s.users[roomCode][id]; ok {
		return true
	}

	return false
}

func (s *Store) AddUser(id string, roomCode string, wsConn *websocket.Conn) {
	if _, ok := s.users[roomCode]; !ok {
		// Need to initialize a new map if it has not been created
		s.users[roomCode] = make(map[string]*UserInfo)
	}
	
	s.users[roomCode][id] = &UserInfo{ hasVoted: false, conn: wsConn }
}

func (s *Store) UpdateUserConn(id string, roomCode string, wsConn *websocket.Conn) {
	s.users[roomCode][id].conn = wsConn
}

func (s *Store) RemoveUser(id string, roomCode string) {
	delete(s.users[roomCode], id)
}

func (s *Store) CastUserVote(id string, roomCode string) {
	prevVal := s.users[roomCode][id].hasVoted
	s.users[roomCode][id].hasVoted = true

	// Only update count if they haven't voted yet
	if prevVal == false {
		log.Println("Vote casted")
		s.voteCount[roomCode]++
	}

	if(s.GetVoteCount(roomCode) > (s.GetTotalUserCount(roomCode)/2)) {
		log.Println("should skip")
		s.notifySkip[roomCode] <- true
		s.resetUsersVote(roomCode)
	}
}

func (s *Store) UserHasVoted(id string, roomCode string) bool {
	if val, ok := s.users[roomCode][id]; ok {
		return val.hasVoted
	}

	return false
}

func (s *Store) resetUsersVote(roomCode string) {
	for id := range s.users[roomCode] {
		s.users[roomCode][id].hasVoted = false	
	}

	s.voteCount[roomCode] = 0
}

func (s *Store) GetTotalUserCount(roomCode string) int {
	return len(s.users[roomCode])
}

func (s *Store) GetVoteCount(roomCode string) int {
	return s.voteCount[roomCode]
}

func (s *Store) IsActiveConn(roomCode string, wsConn *websocket.Conn) bool {
	for _, info := range s.users[roomCode] {
		if info.conn == wsConn {
			return true
		}
	}

	return false
}

func (s *Store) RemoveRoom(roomCode string) {
	delete(s.notifySkip, roomCode)	
	delete(s.users, roomCode)	
}
