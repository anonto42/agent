// Package application holds the chat module's business logic: building the
// model prompt, parsing its reply for a proposed action, gating that action
// through the safety engine, and tracking pending confirmations. The
// interfaces layer (HTTP handlers) only validates input and delegates here.
package application

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/contracts"
)

const systemPrompt = `You are Charli, a helpful, concise browser assistant.

If the user asks you to DO something on the page (fill a field, click a button
or link), respond with ONLY a JSON object, no other text:
  {"action": {"kind": "fill", "target": "<field description>", "value": "<text to enter>"}, "message": "<short confirmation question>"}
  {"action": {"kind": "click", "target": "<button/link text>"}, "message": "<short confirmation question>"}
Otherwise, respond normally in plain text.`

// modelReply is the shape we try to parse out of the LLM's raw text when it
// proposes an action.
type modelReply struct {
	Action  *contracts.Action `json:"action"`
	Message string            `json:"message"`
}

// pendingKey identifies one proposed-but-not-yet-confirmed action.
type pendingKey struct{ session, id string }

// Service is the chat module's business logic.
type Service struct {
	llm     llm.Client
	log     *zap.Logger
	mu      sync.Mutex
	pending map[pendingKey]contracts.Action
}

// NewService constructs the chat application service.
func NewService(client llm.Client, log *zap.Logger) *Service {
	return &Service{llm: client, log: log, pending: make(map[pendingKey]contracts.Action)}
}

// Reply calls the model for a user's message (optionally grounded in page
// text) and returns the event to publish: either a normal answer, or a
// proposed action awaiting confirmation.
func (s *Service) Reply(ctx context.Context, session, id, content, page string) contracts.ChatEvent {
	messages := []llm.Message{{Role: "system", Content: systemPrompt}}
	if page != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "The user is viewing this page:\n" + page,
		})
	}
	messages = append(messages, llm.Message{Role: "user", Content: content})

	raw, err := s.llm.Complete(ctx, messages)
	if err != nil {
		s.log.Warn("llm complete failed", zap.Error(err))
		return contracts.ChatEvent{Type: "error", ID: id, Content: "Charli could not answer right now."}
	}

	action, message, ok := parseAction(raw)
	if !ok {
		return contracts.ChatEvent{Type: "chat", ID: id, Content: raw}
	}

	decision := safety.Evaluate(*action)
	s.log.Info("audit: action proposed",
		zap.String("session", session), zap.String("id", id),
		zap.String("kind", action.Kind), zap.Bool("allowed", decision.Allowed))

	if !decision.Allowed {
		return contracts.ChatEvent{Type: "chat", ID: id, Content: "I can't do that — " + decision.Reason + "."}
	}

	s.mu.Lock()
	s.pending[pendingKey{session, id}] = *action
	s.mu.Unlock()

	if message == "" {
		message = "Should I go ahead?"
	}
	return contracts.ChatEvent{Type: "action", ID: id, Content: message, Action: action}
}

// Confirm resolves a pending action: on approval it re-checks safety (defense
// in depth) and returns an "execute" event; on rejection, "cancelled". Reports
// false if there is no matching pending action (already handled, or unknown).
func (s *Service) Confirm(session, id string, approved bool) (contracts.ChatEvent, bool) {
	key := pendingKey{session, id}

	s.mu.Lock()
	action, found := s.pending[key]
	if found {
		delete(s.pending, key)
	}
	s.mu.Unlock()

	if !found {
		return contracts.ChatEvent{}, false
	}

	if !approved {
		s.log.Info("audit: action rejected", zap.String("session", session), zap.String("id", id))
		return contracts.ChatEvent{Type: "cancelled", ID: id, Content: "Cancelled."}, true
	}

	if decision := safety.Evaluate(action); !decision.Allowed {
		s.log.Warn("audit: action denied on confirm", zap.String("session", session), zap.String("id", id))
		return contracts.ChatEvent{Type: "chat", ID: id, Content: "I can't do that — " + decision.Reason + "."}, true
	}

	s.log.Info("audit: action executed", zap.String("session", session), zap.String("id", id), zap.String("kind", action.Kind))
	return contracts.ChatEvent{Type: "execute", ID: id, Action: &action}, true
}

// parseAction extracts a modelReply's action from raw model text, if present.
// Models sometimes wrap JSON in prose or code fences, so we scan for the first
// balanced {...} block rather than requiring the whole reply to be JSON.
func parseAction(raw string) (*contracts.Action, string, bool) {
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start == -1 || end == -1 || end < start {
		return nil, "", false
	}

	var reply modelReply
	if err := json.Unmarshal([]byte(raw[start:end+1]), &reply); err != nil {
		return nil, "", false
	}
	if reply.Action == nil || reply.Action.Kind == "" {
		return nil, "", false
	}
	return reply.Action, reply.Message, true
}
