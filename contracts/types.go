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
	Session string `json:"session"`        // identifies the caller's SSE stream
	ID      string `json:"id"`             // correlates this message with its reply
	Content string `json:"content"`        // the user's message
	Page    string `json:"page,omitempty"` // text of the page the user is viewing (L1)
}

// Action is something Charli proposes to do on the page (L2). The model selects
// it; the backend safety engine decides whether it runs.
type Action struct {
	Kind   string `json:"kind"`             // "fill" | "click"
	Value  string `json:"value,omitempty"`  // text to type (fill)
	Target string `json:"target,omitempty"` // button/link text to click (click)
}

// ChatEvent is a message pushed down the SSE stream (server -> client).
//
//	type "chat"        a normal assistant answer (content)
//	type "error"       something went wrong (content)
//	type "action"      Charli proposes an action needing confirmation (action set)
//	type "execute"     an approved action to perform on the page (action set)
//	type "cancelled"   a rejected action
//	type "interrupted" the user stopped an in-progress multi-step task (L3)
type ChatEvent struct {
	Type    string  `json:"type"`
	ID      string  `json:"id"` // matches the ChatRequest.ID it answers
	Content string  `json:"content"`
	Action  *Action `json:"action,omitempty"`
}

// ConfirmRequest is the POST /confirm body: the user's decision on a proposed
// action (client -> server).
type ConfirmRequest struct {
	Session  string `json:"session"`
	ID       string `json:"id"`
	Approved bool   `json:"approved"`
}

// ObserveRequest is the POST /observe body (L3): whether an approved action
// actually succeeded when performed on the page, so the agent loop can
// decide its next step (client -> server).
type ObserveRequest struct {
	Session string `json:"session"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Detail  string `json:"detail,omitempty"` // e.g. why it failed
}

// InterruptRequest is the POST /interrupt body (L3): the user's kill switch,
// stopping any in-progress multi-step task (client -> server).
type InterruptRequest struct {
	Session string `json:"session"`
	ID      string `json:"id"`
}
