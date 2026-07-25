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
	// Initialise the sessions map (nil map would panic on write).
	return &Hub{sessions: make(map[string]chan Event)}
}

// Subscribe registers a session and returns its event channel. A prior stream
// for the same session (e.g. a reconnect) is closed and replaced.
func (h *Hub) Subscribe(session string) <-chan Event {
	// Lock for writing: we're mutating the sessions map.
	h.mu.Lock()
	defer h.mu.Unlock()
	// If this session already has a channel, close the old one first.
	if old, ok := h.sessions[session]; ok {
		close(old)
	}
	// Create a new buffered channel and register it.
	ch := make(chan Event, 16)
	h.sessions[session] = ch
	return ch
}

// Unsubscribe removes a session and closes its channel.
func (h *Hub) Unsubscribe(session string) {
	// Lock for writing, then close the channel and remove the session.
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
	// Lock for reading: we only need to look up the channel, not modify the map.
	h.mu.RLock()
	defer h.mu.RUnlock()
	ch, ok := h.sessions[session]
	if !ok {
		return false
	}
	// Try a non-blocking send; drop the event if the buffer is full.
	select {
	case ch <- e:
		return true
	default:
		return false
	}
}
