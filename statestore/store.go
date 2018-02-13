package statestore

import (
	"log"
	"sync"
	"time"
)

// StateStore represents a storage for the state
type StateStore struct {
	states  map[string]string
	updates map[string][]chan string
	mutex   *sync.Mutex
}

// NewStateStore creates a new state store
func NewStateStore() *StateStore {
	return &StateStore{
		states:  map[string]string{},
		updates: map[string][]chan string{},
		mutex:   &sync.Mutex{},
	}
}

// Store saves the given state and returns the old state
func (s *StateStore) Store(name string, state string) (oldState string, changed bool) {
	oldState, _ = s.states[name]
	s.states[name] = state

	if oldState == "" {
		changed = false
	} else {
		changed = oldState != state
	}

	if channels, ok := s.updates[name]; ok {
		for _, ch := range channels {
			ch <- state
		}
	}

	return
}

// Get returns the given state if available, otherwise an empty string
func (s *StateStore) Get(name string) string {
	if s, ok := s.states[name]; ok {
		return s
	}

	return ""
}

// WaitFor waits for the given state to be stored, aborting after the timeout
func (s *StateStore) WaitFor(name string, state string, timeout time.Duration) bool {
	if s.Get(name) == state {
		return true
	}

	after := time.After(timeout)
	updates := s.registerUpdateChannel(name)
	done := false
	result := false

	for !done {
		select {
		case changed := <-updates:
			if changed == state {
				result = true
				done = true
			}
		case <-after:
			log.Printf("Abort waiting for state %s to be %s after %s", name, state, timeout)
			done = true
		}
	}

	s.unregisterUpdateChannel(name, updates)

	return result
}

// WaitForNot waits for the given state to have a different value, aborting after the timeout
func (s *StateStore) WaitForNot(name string, state string, timeout time.Duration) bool {
	if s.Get(name) != state {
		return true
	}

	after := time.After(timeout)
	updates := s.registerUpdateChannel(name)
	done := false
	result := false

	for !done {
		select {
		case changed := <-updates:
			if changed != state {
				result = true
				done = true
			}
		case <-after:
			log.Printf("Abort waiting for state %s not to be %s after %s", name, state, timeout)
			done = true
		}
	}

	s.unregisterUpdateChannel(name, updates)

	return result
}

func (s *StateStore) registerUpdateChannel(name string) chan string {
	ch := make(chan string)

	s.mutex.Lock()
	if _, ok := s.updates[name]; !ok {
		s.updates[name] = []chan string{}
	}
	s.updates[name] = append(s.updates[name], ch)
	s.mutex.Unlock()

	return ch
}

func (s *StateStore) unregisterUpdateChannel(name string, ch chan string) {
	if _, ok := s.updates[name]; !ok {
		return
	}

	channels := []chan string{}

	s.mutex.Lock()
	for _, registeredChannel := range s.updates[name] {
		if registeredChannel != ch {
			channels = append(channels, registeredChannel)
		}
	}
	if len(channels) > 0 {
		s.updates[name] = channels
	} else {
		delete(s.updates, name)
	}
	s.mutex.Unlock()
	close(ch)
}
