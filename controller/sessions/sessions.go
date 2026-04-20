package sessions

import (
	"sync"
)

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string][]string
}

func (s *SessionStore) Add(userID string, pubkey string, maxConcurrent int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check if maxConcurrent has been reached by this user before adding
	if len(s.sessions[userID]) >= maxConcurrent {
		return false
	} else {
		s.sessions[userID] = append(s.sessions[userID], pubkey)
		return true
	}

}

func (s *SessionStore) Remove(userID string, pubkey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.sessions[userID]
	updated := []string{}
	for _, pk := range existing {
		if pk != pubkey { // keep all pubkeys that arent parsed to be removed and store them in updated
			updated = append(updated, pk)
		}
	}

	s.sessions[userID] = updated

}

func (s *SessionStore) Count(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.sessions[userID]

	if !ok {
		return 0
	}

	return len(s.sessions[userID])
}

// constructor
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]string),
	}
}
