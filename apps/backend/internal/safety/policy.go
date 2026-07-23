// Package safety is Charli's policy engine. The model may PROPOSE a tool call,
// but this package makes the final authorization decision — never the model's
// own words. See .agents/shared/rules/agent-safety.md.
package safety

import (
	"strings"

	"github.com/levelaxis/charli/backend/internal/tools"
	"github.com/levelaxis/charli/contracts"
)

// blockedTerms are never allowed as an action's target or value, regardless of
// what the model proposed. Case-insensitive substring match.
var blockedTerms = []string{
	"password",
	"credit card",
	"card number",
	"cvv",
	"ssn",
	"social security",
	"bank account",
	"delete account",
	"wire transfer",
}

// Decision is the outcome of evaluating a proposed action.
type Decision struct {
	Allowed bool
	Reason  string // set when Allowed is false
}

// Engine evaluates proposed actions against the tool registry and Charli's
// content policy.
type Engine struct {
	registry *tools.Registry
}

// NewEngine builds an Engine backed by the given tool registry.
func NewEngine(registry *tools.Registry) *Engine {
	return &Engine{registry: registry}
}

// Evaluate decides whether a proposed action may even be shown to the user for
// confirmation. L2 v1 policy: every action requires user confirmation (nothing
// auto-executes yet); this function only handles the hard "never" cases.
func (e *Engine) Evaluate(action contracts.Action) Decision {
	tool, ok := e.registry.Lookup(action.Kind)
	if !ok {
		return Decision{Allowed: false, Reason: "unknown action kind"}
	}
	if err := tool.Validate(action); err != nil {
		return Decision{Allowed: false, Reason: err.Error()}
	}

	haystack := strings.ToLower(action.Target + " " + action.Value)
	for _, term := range blockedTerms {
		if strings.Contains(haystack, term) {
			return Decision{Allowed: false, Reason: "this looks like a sensitive field (" + term + "); Charli won't touch it"}
		}
	}

	return Decision{Allowed: true}
}
