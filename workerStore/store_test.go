package workerStore

import (
	"testing"
)

func TestAddPoller(t *testing.T) {
	roomCode := "11111"
	expected := AddPoller(roomCode)
	actual := GetPollerChan(roomCode) 

	if actual != expected {
		t.Errorf("got %p, want %p", actual, expected)
	}
}

func TestRemovePoller(t *testing.T) {
	roomCode := "22222"
	AddPoller(roomCode)
	RemovePoller(roomCode)
	actual := GetPollerChan(roomCode) 

	if actual != nil {
		t.Errorf("got %p, want nil", actual)
	}
}
