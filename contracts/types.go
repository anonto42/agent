// Package contracts is the SINGLE SOURCE OF TRUTH for the data shapes shared
// between the Go backend and the TypeScript extension/website.
//
// You define types here once, in Go. Running `moon run contracts:generate`
// (tygo) emits the matching TypeScript into packages/shared/src/types.gen.ts,
// so the two languages can never drift apart.
//
// Transport: chat is SSE (server -> client, GET /events) + POST (client ->
// server, POST /chat). The types below describe that protocol.
package contracts

// AgentRole identifies who produced a message in a conversation.
type AgentRole string

const (
	RoleUser      AgentRole = "user"
	RoleAssistant AgentRole = "assistant"
	RoleSystem    AgentRole = "system"
)

// ChatMessage is one turn in a Charli conversation (the UI's model).
type ChatMessage struct {
	Role    AgentRole `json:"role"`
	Content string    `json:"content"`
}

// ChatRequest is the POST /chat body sent by the client (client -> server).
type ChatRequest struct {
	Session string `json:"session"` // identifies the caller's SSE stream
	ID      string `json:"id"`      // correlates this message with its reply
	Content string `json:"content"`
}

// ChatEvent is a message pushed down the SSE stream (server -> client).
type ChatEvent struct {
	Type    string `json:"type"`    // "chat" | "error"
	ID      string `json:"id"`      // matches the ChatRequest.ID it answers
	Content string `json:"content"`
}
