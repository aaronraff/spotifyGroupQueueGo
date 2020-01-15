package workerStore

var store = make(map[string]chan bool)

// AddPoller creates and returns a buffered channel, which is primarily used to
// signal to a room's poller that it should be stopped. It also stores this
// channel in a map (called store) which maps roomCodes to their cancellation 
// channels.
func AddPoller(roomCode string) chan bool {
	// Make buffered so it doesn't block when closing the room
	cancelChan := make(chan bool, 1)
	store[roomCode] = cancelChan
	
	return cancelChan
}

// RemovePoller removes the entry for the specified roomCode from the store.
func RemovePoller(roomCode string) {
	delete(store, roomCode)
}

// GetPollerChan returns the chan for the specified roomCode.
func GetPollerChan(roomCode string) chan bool {
	return store[roomCode]
}
