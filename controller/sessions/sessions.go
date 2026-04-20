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

// Removed USERID as this will not be sent from node json
func (s *SessionStore) Remove(pubkey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// find which user owns pubkey (need USERID to be map key for count )
	for userID, pubkeys := range s.sessions {
		for i, pk := range pubkeys {
			if pk == pubkey {
				s.sessions[userID] = append(pubkeys[:i], pubkeys[i+1:]...)
				return
			}
		}
	}
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

// func (s *SessionStore) FindUser(pubkey string) (string, bool) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	for userID, pubkeys := range s.sessions {
// 		for _, pk := range pubkeys {
// 			if pk == pubkey {
// 				return userID, true
// 			}
// 		}
// 	}
// 	return "", false
// }

// constructor
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]string),
	}
}
