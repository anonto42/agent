package application

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
)

type stubLLM struct{ reply string }

func (s stubLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	return s.reply, nil
}

func TestReplyPlainText(t *testing.T) {
	svc := NewService(stubLLM{reply: "hello there"}, zap.NewNop())
	event := svc.Reply(context.Background(), "s1", "1", "hi", "")
	if event.Type != "chat" || event.Content != "hello there" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestReplyProposesAction(t *testing.T) {
	reply := `{"action":{"kind":"fill","target":"email field","value":"me@example.com"},"message":"Fill in your email?"}`
	svc := NewService(stubLLM{reply: reply}, zap.NewNop())

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
	svc := NewService(stubLLM{reply: reply}, zap.NewNop())

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
	svc := NewService(stubLLM{reply: reply}, zap.NewNop())
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
	svc := NewService(stubLLM{reply: reply}, zap.NewNop())
	svc.Reply(context.Background(), "s1", "1", "leave a comment", "")

	event, found := svc.Confirm("s1", "1", false)
	if !found || event.Type != "cancelled" {
		t.Fatalf("unexpected reject result: found=%v event=%+v", found, event)
	}
}

func TestConfirmUnknownIDNotFound(t *testing.T) {
	svc := NewService(stubLLM{reply: "n/a"}, zap.NewNop())
	if _, found := svc.Confirm("s1", "does-not-exist", true); found {
		t.Fatal("expected no pending action for an unknown id")
	}
}
