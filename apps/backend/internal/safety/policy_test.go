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

// TestEvaluateRequiresConfirmationPerRiskTier locks in agent-safety.md's risk
// tiers: fill (RiskAuto) may run on its own, click (RiskConfirm) may not.
func TestEvaluateRequiresConfirmationPerRiskTier(t *testing.T) {
	engine := NewEngine(tools.Default())

	fill := engine.Evaluate(contracts.Action{Kind: "fill", Target: "email field", Value: "me@example.com"})
	if !fill.Allowed || fill.RequiresConfirmation {
		t.Errorf("expected fill to be auto-allowed without confirmation, got %+v", fill)
	}

	click := engine.Evaluate(contracts.Action{Kind: "click", Target: "Submit"})
	if !click.Allowed || !click.RequiresConfirmation {
		t.Errorf("expected click to require confirmation, got %+v", click)
	}
}

// TestEvaluateBlocksRiskBlockTools proves a RiskBlock-tier tool is denied
// outright, even though no default tool currently uses that tier.
func TestEvaluateBlocksRiskBlockTools(t *testing.T) {
	registry := tools.NewRegistry(tools.Tool{
		Kind:          "wire_transfer",
		Risk:          tools.RiskBlock,
		PromptExample: `{"kind": "wire_transfer"}`,
		Validate:      func(contracts.Action) error { return nil },
	})
	engine := NewEngine(registry)

	if d := engine.Evaluate(contracts.Action{Kind: "wire_transfer"}); d.Allowed {
		t.Errorf("expected a RiskBlock tool to be denied, got %+v", d)
	}
}
