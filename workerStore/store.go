package workerStore

var store = make(map[string]chan bool)

func AddPoller(roomCode string) chan bool {
	// Make buffered so it doesn't block when closing the room
	cancelChan := make(chan bool, 1)
	store[roomCode] = cancelChan
	
	return cancelChan
}

func RemovePoller(roomCode string) {
	delete(store, roomCode)
}

func GetPollerChan(roomCode string) chan bool {
	return store[roomCode]
}
