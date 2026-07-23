package safety

import (
	"testing"

	"github.com/levelaxis/charli/backend/internal/tools"
	"github.com/levelaxis/charli/contracts"
)

func TestEvaluateAllowsOrdinaryActions(t *testing.T) {
	engine := NewEngine(tools.Default())
	for _, a := range []contracts.Action{
		{Kind: "fill", Target: "email field", Value: "me@example.com"},
		{Kind: "click", Target: "Submit"},
	} {
		if d := engine.Evaluate(a); !d.Allowed {
			t.Errorf("expected %+v to be allowed, got denied: %s", a, d.Reason)
		}
	}
}

func TestEvaluateBlocksSensitiveFields(t *testing.T) {
	engine := NewEngine(tools.Default())
	for _, a := range []contracts.Action{
		{Kind: "fill", Target: "password field", Value: "hunter2"},
		{Kind: "fill", Target: "card number", Value: "4111111111111111"},
		{Kind: "click", Target: "Delete Account"},
	} {
		if d := engine.Evaluate(a); d.Allowed {
			t.Errorf("expected %+v to be denied, but it was allowed", a)
		}
	}
}

func TestEvaluateRejectsUnknownKind(t *testing.T) {
	engine := NewEngine(tools.Default())
	if d := engine.Evaluate(contracts.Action{Kind: "delete", Target: "row 1"}); d.Allowed {
		t.Error("unknown action kind should be denied")
	}
}

func TestEvaluateRejectsMalformedArgs(t *testing.T) {
	engine := NewEngine(tools.Default())
	for _, a := range []contracts.Action{
		{Kind: "fill", Target: "email field", Value: ""},
		{Kind: "click", Target: ""},
	} {
		if d := engine.Evaluate(a); d.Allowed {
			t.Errorf("expected %+v to be denied for missing required arg, but it was allowed", a)
		}
	}
}
