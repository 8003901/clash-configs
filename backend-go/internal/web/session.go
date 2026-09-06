package web

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

const sessionCookie = "clash_session"

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string // sessionID -> username
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]string{}}
}

func (s *SessionStore) Create(username string) string {
	id := newID()
	s.mu.Lock()
	s.sessions[id] = username
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(id string) (string, bool) {
	s.mu.RLock()
	u, ok := s.sessions[id]
	s.mu.RUnlock()
	return u, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
