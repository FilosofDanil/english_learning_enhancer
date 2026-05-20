package session

import (
	"sync"

	"english_learn_tg_bot/internal/content"
)

// QuizSession mirrors HOW_TO_CHECK (direction fixed DE→EN in production).
type QuizSession struct {
	UserID  int64
	Phrases []content.Phrase
	Current int
	Score   float64
}

// Store abstracts session persistence (in-memory impl by default).
type Store interface {
	Start(userID int64, phrases []content.Phrase) *QuizSession
	Get(userID int64) (*QuizSession, bool)
	Remove(userID int64)
}

// Memory is thread-safe in-memory quiz storage.
type Memory struct {
	mu   sync.RWMutex
	data map[int64]*QuizSession
}

// NewMemory creates an empty Memory store.
func NewMemory() *Memory {
	return &Memory{data: make(map[int64]*QuizSession)}
}

// Start replaces any session for userID with a new one pointing at phrases.
func (m *Memory) Start(userID int64, phrases []content.Phrase) *QuizSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := &QuizSession{
		UserID:  userID,
		Phrases: phrases,
		Current: 0,
		Score:   0,
	}
	m.data[userID] = s
	return s
}

// Get returns active session when present.
func (m *Memory) Get(userID int64) (*QuizSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.data[userID]
	return s, ok && s != nil
}

// Remove drops the session key.
func (m *Memory) Remove(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, userID)
}
