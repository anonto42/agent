// Package stream provides an in-memory pub/sub used to push realtime events to
// clients over Server-Sent Events (SSE). One channel per session.
package stream

import (
	"sync"

	"github.com/levelaxis/charli/contracts"
)

// Event is a single message pushed down a client's stream — the same shape the
// extension receives, so this is a type alias to the shared contract (the
// single source of truth), not a hand-copied struct.
type Event = contracts.ChatEvent

// Hub tracks active client sessions and fans events out to them.
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]chan Event
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{sessions: make(map[string]chan Event)}
}

// Subscribe registers a session and returns its event channel. A prior stream
// for the same session (e.g. a reconnect) is closed and replaced.
func (h *Hub) Subscribe(session string) <-chan Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.sessions[session]; ok {
		close(old)
	}
	ch := make(chan Event, 16)
	h.sessions[session] = ch
	return ch
}

// Unsubscribe removes a session and closes its channel.
func (h *Hub) Unsubscribe(session string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, ok := h.sessions[session]; ok {
		close(ch)
		delete(h.sessions, session)
	}
}

// Publish delivers an event to a session. It reports whether a live session
// received it. A full buffer drops the event (Phase 0 behaviour).
func (h *Hub) Publish(session string, e Event) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ch, ok := h.sessions[session]
	if !ok {
		return false
	}
	select {
	case ch <- e:
		return true
	default:
		return false
	}
}
