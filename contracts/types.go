// Package contracts is the SINGLE SOURCE OF TRUTH for the data shapes
// shared between the Go backend and the TypeScript extension/website.
//
// You define types here once, in Go. Running `moon run contracts:generate`
// (tygo) emits the matching TypeScript into packages/shared/src/types.gen.ts,
// so the two languages can never drift apart.
package contracts

// AgentRole identifies who produced a message in a conversation.
type AgentRole string

const (
	RoleUser      AgentRole = "user"
	RoleAssistant AgentRole = "assistant"
	RoleSystem    AgentRole = "system"
)

// ChatMessage is one turn in a Charli conversation.
type ChatMessage struct {
	Role    AgentRole `json:"role"`
	Content string    `json:"content"`
}

// WSMessage is the envelope for every websocket frame exchanged
// between the extension and the backend.
type WSMessage struct {
	Type    string `json:"type"`    // e.g. "chat", "action", "ping"
	Payload any    `json:"payload"` // shape depends on Type
}
