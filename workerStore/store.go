package workerStore

var store = make(map[string]chan bool)

func AddPoller(roomCode string) chan bool {
	cancelChan := make(chan bool)
	store[roomCode] = cancelChan
	
	return cancelChan
}

func RemovePoller(roomCode string) {
	delete(store, roomCode)
}

func GetPollerChan(roomCode string) chan bool {
	return store[roomCode]
}
