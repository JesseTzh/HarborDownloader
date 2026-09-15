package credential

import "sync"

// Store keeps Harbor credentials in process memory for the current session.
// Disk persistence is handled by config.Service; this store is never logged.
type Store struct {
	mu       sync.Mutex
	registry string
	username string
	password string
}

func New() *Store {
	return &Store{}
}

func (s *Store) Set(registry, username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry = registry
	s.username = username
	s.password = password
}

func (s *Store) Get(registry string) (username, password string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.registry == "" || s.registry != registry {
		return "", "", false
	}
	if s.username == "" && s.password == "" {
		return "", "", false
	}
	return s.username, s.password, true
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry = ""
	s.username = ""
	s.password = ""
}
