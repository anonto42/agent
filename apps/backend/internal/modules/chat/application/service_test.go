package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/tools"
	"github.com/levelaxis/charli/contracts"
)

type stubLLM struct{ reply string }

func (s stubLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	return s.reply, nil
}

// sequenceLLM returns one reply per call, in the given order; calls past the
// end of the slice repeat the last reply.
type sequenceLLM struct {
	replies []string
	calls   int
}

func (s *sequenceLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	i := s.calls
	if i >= len(s.replies) {
		i = len(s.replies) - 1
	}
	s.calls++
	return s.replies[i], nil
}

// alwaysProposeLLM always proposes the same action — used to drive a loop
// into its turn limit.
type alwaysProposeLLM struct{}

func (alwaysProposeLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	return `{"action":{"kind":"click","target":"Next"},"message":"Next?"}`, nil
}

// blockingLLM never returns on its own — it blocks until its context is
// cancelled, so tests can prove an interrupt aborts an in-flight call rather
// than just refusing future ones.
type blockingLLM struct{ started chan struct{} }

func (b *blockingLLM) Complete(ctx context.Context, _ []llm.Message) (string, error) {
	close(b.started)
	<-ctx.Done()
	return "", ctx.Err()
}

// newTestService wires a Service with the default tool registry and a safety
// engine backed by it — the same wiring internal/app performs.
func newTestService(client llm.Client) *Service {
	registry := tools.Default()
	return NewService(client, zap.NewNop(), registry, safety.NewEngine(registry))
}

func TestReplyPlainText(t *testing.T) {
	svc := newTestService(stubLLM{reply: "hello there"})
	event := svc.Reply(context.Background(), "s1", "1", "hi", "")
	if event.Type != "chat" || event.Content != "hello there" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestReplyProposesAction(t *testing.T) {
	reply := `{"action":{"kind":"fill","target":"email field","value":"me@example.com"},"message":"Fill in your email?"}`
	svc := newTestService(stubLLM{reply: reply})

	event := svc.Reply(context.Background(), "s1", "1", "fill in my email", "")
	if event.Type != "action" {
		t.Fatalf("expected action event, got %+v", event)
	}
	if event.Action == nil || event.Action.Kind != "fill" || event.Action.Value != "me@example.com" {
		t.Fatalf("unexpected action: %+v", event.Action)
	}
	if event.Content != "Fill in your email?" {
		t.Fatalf("unexpected message: %q", event.Content)
	}
}

func TestReplyDeniesSensitiveAction(t *testing.T) {
	reply := `{"action":{"kind":"fill","target":"password field","value":"hunter2"},"message":"ok?"}`
	svc := newTestService(stubLLM{reply: reply})

	event := svc.Reply(context.Background(), "s1", "1", "fill my password", "")
	if event.Type != "chat" {
		t.Fatalf("expected a denial as a plain chat event, got %+v", event)
	}
	if event.Action != nil {
		t.Fatalf("denied action must not be attached to the event: %+v", event.Action)
	}

	// And critically: nothing pending, so confirming it must fail.
	if _, found := svc.Confirm("s1", "1", true); found {
		t.Fatal("a denied action must never become confirmable")
	}
}

func TestConfirmApprovedReturnsExecute(t *testing.T) {
	reply := `{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`
	svc := newTestService(stubLLM{reply: reply})
	svc.Reply(context.Background(), "s1", "1", "click submit", "")

	event, found := svc.Confirm("s1", "1", true)
	if !found {
		t.Fatal("expected a pending action to be found")
	}
	if event.Type != "execute" || event.Action == nil || event.Action.Kind != "click" {
		t.Fatalf("unexpected confirm result: %+v", event)
	}

	// Confirming again must fail — it was consumed.
	if _, found := svc.Confirm("s1", "1", true); found {
		t.Fatal("a pending action must only be confirmable once")
	}
}

func TestConfirmRejectedReturnsCancelled(t *testing.T) {
	reply := `{"action":{"kind":"fill","target":"comment box","value":"hi"},"message":"ok?"}`
	svc := newTestService(stubLLM{reply: reply})
	svc.Reply(context.Background(), "s1", "1", "leave a comment", "")

	event, found := svc.Confirm("s1", "1", false)
	if !found || event.Type != "cancelled" {
		t.Fatalf("unexpected reject result: found=%v event=%+v", found, event)
	}
}

func TestConfirmUnknownIDNotFound(t *testing.T) {
	svc := newTestService(stubLLM{reply: "n/a"})
	if _, found := svc.Confirm("s1", "does-not-exist", true); found {
		t.Fatal("expected no pending action for an unknown id")
	}
}

// TestObserveContinuesTaskThenFinishes exercises the L3 loop end to end:
// propose -> confirm -> observe(success) -> the model's next turn is a plain
// final answer, ending the task.
func TestObserveContinuesTaskThenFinishes(t *testing.T) {
	client := &sequenceLLM{replies: []string{
		`{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`,
		"All done!",
	}}
	svc := newTestService(client)

	event := svc.Reply(context.Background(), "s1", "1", "click submit", "")
	if event.Type != "action" {
		t.Fatalf("expected action event, got %+v", event)
	}

	confirmEvent, found := svc.Confirm("s1", "1", true)
	if !found || confirmEvent.Type != "execute" {
		t.Fatalf("unexpected confirm result: found=%v event=%+v", found, confirmEvent)
	}

	final := svc.Observe(context.Background(), "s1", "1", true, "")
	if final.Type != "chat" || final.Content != "All done!" {
		t.Fatalf("expected the loop to continue and finish, got %+v", final)
	}

	// The task is over: neither confirming nor observing it again finds anything.
	if _, found := svc.Confirm("s1", "1", true); found {
		t.Fatal("a finished task must not still have a pending action")
	}
	if empty := svc.Observe(context.Background(), "s1", "1", true, ""); empty.Type != "" {
		t.Fatalf("observing a finished task must be a no-op, got %+v", empty)
	}
}

// TestLoopStopsAtMaxTurns proves the loop bounds itself even when the model
// never stops proposing actions.
func TestLoopStopsAtMaxTurns(t *testing.T) {
	svc := newTestService(alwaysProposeLLM{})
	svc.maxTurns = 2 // keep the test fast

	event := svc.Reply(context.Background(), "s1", "1", "keep clicking", "")
	for i := 0; i < 10 && event.Type == "action"; i++ {
		confirmEvent, found := svc.Confirm("s1", "1", true)
		if !found || confirmEvent.Type != "execute" {
			t.Fatalf("unexpected confirm result: found=%v event=%+v", found, confirmEvent)
		}
		event = svc.Observe(context.Background(), "s1", "1", true, "")
	}

	if event.Type != "chat" || !strings.Contains(event.Content, "step limit") {
		t.Fatalf("expected the loop to stop at the turn limit, got %+v", event)
	}
}

// TestInterruptStopsPendingTask covers the common kill-switch case: stopping
// a task while it's waiting on the user's confirmation.
func TestInterruptStopsPendingTask(t *testing.T) {
	svc := newTestService(stubLLM{reply: `{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`})
	svc.Reply(context.Background(), "s1", "1", "click submit", "")

	event, found := svc.Interrupt("s1", "1")
	if !found || event.Type != "interrupted" {
		t.Fatalf("unexpected interrupt result: found=%v event=%+v", found, event)
	}

	if _, found := svc.Confirm("s1", "1", true); found {
		t.Fatal("an interrupted task must not still have a pending action")
	}
	if _, found := svc.Interrupt("s1", "1"); found {
		t.Fatal("interrupting an already-interrupted task should find nothing")
	}
}

// TestInterruptCancelsInFlightLLMCall proves the kill switch aborts a call
// that's already in flight, and that the stale response it eventually
// produces is silently discarded rather than reviving the interrupted task.
func TestInterruptCancelsInFlightLLMCall(t *testing.T) {
	client := &blockingLLM{started: make(chan struct{})}
	svc := newTestService(client)

	done := make(chan contracts.ChatEvent, 1)
	go func() {
		done <- svc.Reply(context.Background(), "s1", "1", "do something slow", "")
	}()

	<-client.started // the LLM call is now actually in flight

	event, found := svc.Interrupt("s1", "1")
	if !found || event.Type != "interrupted" {
		t.Fatalf("unexpected interrupt result: found=%v event=%+v", found, event)
	}

	select {
	case replyEvent := <-done:
		if replyEvent.Type != "" {
			t.Fatalf("a stale in-flight reply must not publish anything, got %+v", replyEvent)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Reply did not return after being interrupted")
	}
}
