package session

import "sync"

// UserRegistry tracks chat IDs of everyone who has interacted with the bot,
// so the admin can broadcast to them later.
type UserRegistry interface {
	Add(userID, chatID int64)
	ChatIDs() []int64
}

// MemoryRegistry is a thread-safe in-memory user registry.
type MemoryRegistry struct {
	mu   sync.RWMutex
	data map[int64]int64 // userID -> chatID
}

// NewMemoryRegistry creates an empty registry.
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{data: make(map[int64]int64)}
}

// Add records (or updates) the chat ID for a user.
func (r *MemoryRegistry) Add(userID, chatID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[userID] = chatID
}

// ChatIDs returns a snapshot of all known chat IDs.
func (r *MemoryRegistry) ChatIDs() []int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]int64, 0, len(r.data))
	for _, chatID := range r.data {
		out = append(out, chatID)
	}
	return out
}
