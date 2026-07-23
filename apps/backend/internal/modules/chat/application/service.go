// Package application holds the chat module's business logic: building the
// model prompt, running the multi-step agent loop (L3) — parsing a proposed
// action, gating it through the safety engine, tracking the pending
// confirmation, applying the executed result, and re-deciding — until the
// model gives a final answer, the loop hits its turn limit, an action is
// denied, or the user interrupts. The interfaces layer (HTTP handlers) only
// validates input and delegates here.
package application

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/tools"
	"github.com/levelaxis/charli/contracts"
)

// defaultMaxTurns bounds every agent loop: after this many proposed actions
// in one task, the loop stops itself rather than run indefinitely.
const defaultMaxTurns = 6

// buildSystemPrompt renders the model's instructions, including one example
// per registered tool, so adding a tool never requires editing this prompt
// by hand.
func buildSystemPrompt(registry *tools.Registry) string {
	var b strings.Builder
	b.WriteString("You are Charli, a helpful, concise browser assistant.\n\n")
	b.WriteString("If the user asks you to DO something on the page (fill a field, click a button\n")
	b.WriteString("or link), respond with a single JSON object, no other text:\n")
	for _, t := range registry.All() {
		b.WriteString(`  {"action": ` + t.PromptExample + `, "message": "<short confirmation question>"}` + "\n")
	}
	b.WriteString("Return exactly one action at a time. Never return multiple actions in one response.\n")
	b.WriteString("After you propose an action, you'll be told whether it succeeded before you\n")
	b.WriteString("continue — use that to decide your next step, or to give a final plain-text answer\n")
	b.WriteString("once the task is done.\n")
	b.WriteString("Otherwise, respond normally in plain text.")
	return b.String()
}

// modelReply is the shape we try to parse out of the LLM's raw text when it
// proposes an action.
type modelReply struct {
	Action  *contracts.Action `json:"action"`
	Message string            `json:"message"`
}

// pendingKey identifies one proposed-but-not-yet-confirmed action.
type pendingKey struct{ session, id string }

// taskState is one in-progress multi-step task (L3) for a session. Charli
// runs at most one task per session at a time, matching the UI (the
// composer is disabled for the whole task, not just per step).
type taskState struct {
	id       string
	messages []llm.Message
	turn     int
	cancel   context.CancelFunc // aborts the in-flight LLM call, if any
}

// Service is the chat module's business logic.
type Service struct {
	llm          llm.Client
	log          *zap.Logger
	safety       *safety.Engine
	systemPrompt string
	maxTurns     int
	mu           sync.Mutex
	pending      map[pendingKey]contracts.Action
	tasks        map[string]*taskState // keyed by session
}

// NewService constructs the chat application service.
func NewService(client llm.Client, log *zap.Logger, registry *tools.Registry, safetyEngine *safety.Engine) *Service {
	return &Service{
		llm:          client,
		log:          log,
		safety:       safetyEngine,
		systemPrompt: buildSystemPrompt(registry),
		maxTurns:     defaultMaxTurns,
		pending:      make(map[pendingKey]contracts.Action),
		tasks:        make(map[string]*taskState),
	}
}

// Reply starts a new task for a user's message (optionally grounded in page
// text) and runs its first turn. The returned event is either a normal
// answer (the task is already done) or a proposed action awaiting
// confirmation (the task continues once that action is observed).
func (s *Service) Reply(ctx context.Context, session, id, content, page string) contracts.ChatEvent {
	messages := []llm.Message{{Role: "system", Content: s.systemPrompt}}
	if page != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "The user is viewing this page:\n" + page,
		})
	}
	messages = append(messages, llm.Message{Role: "user", Content: content})

	state := &taskState{id: id, messages: messages}
	s.mu.Lock()
	s.tasks[session] = state
	s.mu.Unlock()

	return s.step(ctx, session, state)
}

// step is one turn of the agent loop: ask the model, and either end the task
// (plain answer, denial, or turn-limit) or propose the next action. Returns
// a zero-value ChatEvent if the task was interrupted while this turn's LLM
// call was in flight — the interrupt already published the authoritative
// outcome, so the caller must not publish anything for a zero-value event.
func (s *Service) step(ctx context.Context, session string, state *taskState) contracts.ChatEvent {
	id := state.id

	if state.turn >= s.maxTurns {
		s.clearTaskIfCurrent(session, state)
		return contracts.ChatEvent{Type: "chat", ID: id, Content: "I've reached my step limit for this task, so I'll stop here."}
	}

	stepCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	state.cancel = cancel
	s.mu.Unlock()
	defer cancel()

	raw, err := s.llm.Complete(stepCtx, state.messages)

	if !s.isCurrent(session, state) {
		// Interrupted mid-call; /interrupt already published the outcome.
		return contracts.ChatEvent{}
	}

	if err != nil {
		s.log.Warn("llm complete failed", zap.Error(err))
		s.clearTaskIfCurrent(session, state)
		return contracts.ChatEvent{Type: "error", ID: id, Content: "Charli could not answer right now."}
	}

	action, message, ok := parseAction(raw)
	if !ok {
		s.clearTaskIfCurrent(session, state)
		return contracts.ChatEvent{Type: "chat", ID: id, Content: raw}
	}

	decision := s.safety.Evaluate(*action)
	s.log.Info("audit: action proposed",
		zap.String("session", session), zap.String("id", id),
		zap.String("kind", action.Kind), zap.Bool("allowed", decision.Allowed), zap.Int("turn", state.turn))

	if !decision.Allowed {
		s.clearTaskIfCurrent(session, state)
		return contracts.ChatEvent{Type: "chat", ID: id, Content: "I can't do that — " + decision.Reason + "."}
	}

	state.messages = append(state.messages, llm.Message{Role: "assistant", Content: raw})

	s.mu.Lock()
	s.pending[pendingKey{session, id}] = *action
	s.mu.Unlock()

	if message == "" {
		message = "Should I go ahead?"
	}
	return contracts.ChatEvent{Type: "action", ID: id, Content: message, Action: action}
}

// Confirm resolves a pending action: on approval it re-checks safety (defense
// in depth) and returns an "execute" event — the task continues once the
// caller reports back via Observe. On rejection, the whole task ends.
// Reports false if there is no matching pending action (already handled, or
// unknown).
func (s *Service) Confirm(session, id string, approved bool) (contracts.ChatEvent, bool) {
	key := pendingKey{session, id}

	s.mu.Lock()
	action, found := s.pending[key]
	if found {
		delete(s.pending, key)
	}
	state := s.tasks[session]
	s.mu.Unlock()

	if !found {
		return contracts.ChatEvent{}, false
	}

	if !approved {
		s.clearTaskIfCurrent(session, state)
		s.log.Info("audit: action rejected", zap.String("session", session), zap.String("id", id))
		return contracts.ChatEvent{Type: "cancelled", ID: id, Content: "Cancelled."}, true
	}

	if decision := s.safety.Evaluate(action); !decision.Allowed {
		s.clearTaskIfCurrent(session, state)
		s.log.Warn("audit: action denied on confirm", zap.String("session", session), zap.String("id", id))
		return contracts.ChatEvent{Type: "chat", ID: id, Content: "I can't do that — " + decision.Reason + "."}, true
	}

	s.log.Info("audit: action executed", zap.String("session", session), zap.String("id", id), zap.String("kind", action.Kind))
	return contracts.ChatEvent{Type: "execute", ID: id, Action: &action}, true
}

// HasTask reports whether session has an in-progress task matching id — a
// cheap synchronous check the handler uses to reject an unknown/stale
// observation before spawning the goroutine that would call the model.
func (s *Service) HasTask(session, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.tasks[session]
	return ok && state.id == id
}

// Observe applies the result of an executed action and runs the next turn of
// the loop. Returns a zero-value ChatEvent if there's no matching in-progress
// task (already finished, interrupted, or unknown) — the caller must not
// publish anything for a zero-value event.
func (s *Service) Observe(ctx context.Context, session, id string, success bool, detail string) contracts.ChatEvent {
	s.mu.Lock()
	state, ok := s.tasks[session]
	if !ok || state.id != id {
		s.mu.Unlock()
		return contracts.ChatEvent{}
	}
	state.turn++
	s.mu.Unlock()

	observation := "The action succeeded."
	if !success {
		observation = "The action failed"
		if detail != "" {
			observation += ": " + detail
		}
		observation += "."
	}
	state.messages = append(state.messages, llm.Message{Role: "system", Content: "Observation: " + observation})

	return s.step(ctx, session, state)
}

// Interrupt is the user's kill switch: it stops session's in-progress task
// (cancelling an in-flight LLM call, if any) and discards any pending
// confirmation. Reports false if id doesn't match the active task.
func (s *Service) Interrupt(session, id string) (contracts.ChatEvent, bool) {
	s.mu.Lock()
	state, ok := s.tasks[session]
	matches := ok && state.id == id
	if matches {
		delete(s.tasks, session)
		delete(s.pending, pendingKey{session, id})
	}
	s.mu.Unlock()

	if !matches {
		return contracts.ChatEvent{}, false
	}

	if state.cancel != nil {
		state.cancel()
	}

	s.log.Info("audit: task interrupted", zap.String("session", session), zap.String("id", id))
	return contracts.ChatEvent{Type: "interrupted", ID: id, Content: "Stopped."}, true
}

// isCurrent reports whether state is still the active task for session (it
// may have been replaced or removed by an interrupt or a new Reply).
func (s *Service) isCurrent(session string, state *taskState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tasks[session] == state
}

// clearTaskIfCurrent removes state from session's slot, but only if it's
// still there — avoids clobbering a newer task that may have replaced it.
func (s *Service) clearTaskIfCurrent(session string, state *taskState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tasks[session] == state {
		delete(s.tasks, session)
	}
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
