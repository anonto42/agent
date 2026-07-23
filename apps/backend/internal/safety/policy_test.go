package safety

import (
	"testing"

	"github.com/levelaxis/charli/contracts"
)

func TestEvaluateAllowsOrdinaryActions(t *testing.T) {
	for _, a := range []contracts.Action{
		{Kind: "fill", Target: "email field", Value: "me@example.com"},
		{Kind: "click", Target: "Submit"},
	} {
		if d := Evaluate(a); !d.Allowed {
			t.Errorf("expected %+v to be allowed, got denied: %s", a, d.Reason)
		}
	}
}

func TestEvaluateBlocksSensitiveFields(t *testing.T) {
	for _, a := range []contracts.Action{
		{Kind: "fill", Target: "password field", Value: "hunter2"},
		{Kind: "fill", Target: "card number", Value: "4111111111111111"},
		{Kind: "click", Target: "Delete Account"},
	} {
		if d := Evaluate(a); d.Allowed {
			t.Errorf("expected %+v to be denied, but it was allowed", a)
		}
	}
}

func TestEvaluateRejectsUnknownKind(t *testing.T) {
	if d := Evaluate(contracts.Action{Kind: "delete", Target: "row 1"}); d.Allowed {
		t.Error("unknown action kind should be denied")
	}
}
