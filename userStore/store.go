package userStore

type Store struct {
	users map[string]map[string]bool
	voteCount map[string]int
}

func NewStore() *Store {	
	return &Store { users: make(map[string]map[string]bool), voteCount: make(map[string]int) }
}

func (s *Store) AddUser(id string, roomCode string) {
	s.users[roomCode][id] = false
}

func (s *Store) RemoveUser(id string, roomCode string) {
	delete(s.users[roomCode], id)
}

func (s *Store) UserVoted(id string, roomCode string) {
	s.users[roomCode][id] = true
	s.voteCount[roomCode]++
}

func (s *Store) ResetUsersVote(roomCode string) {
	for id := range s.users[roomCode] {
		s.users[roomCode][id] = false	
	}

	s.voteCount[roomCode] = 0
}

func (s *Store) GetTotalUserCount(roomCode string) int {
	return len(s.users[roomCode])
}

func (s *Store) GetVoteCount(roomCode string) int {
	return s.voteCount[roomCode]
}
