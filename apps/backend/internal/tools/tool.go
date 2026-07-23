// Package tools is the registry of actions Charli's model may propose. It is
// the single place that knows every valid action kind, its argument shape,
// and its intended risk tier — apps/backend/internal/safety validates
// against it, and apps/backend/internal/modules/chat/application builds the
// model's system prompt from it, so neither has to hardcode kinds itself.
package tools

import "github.com/levelaxis/charli/contracts"

// RiskTier is a tool's intended confirmation policy, per
// .agents/shared/rules/agent-safety.md. It is captured here as metadata only;
// nothing reads it yet — every proposed action still requires user
// confirmation regardless of tier (see internal/safety.Engine.Evaluate).
type RiskTier string

const (
	RiskAuto    RiskTier = "auto"    // read-only / reversible, no confirmation
	RiskConfirm RiskTier = "confirm" // needs explicit user approval
	RiskBlock   RiskTier = "block"   // never allowed
)

// Tool describes one action kind the model can propose.
type Tool struct {
	Kind string
	Risk RiskTier

	// PromptExample is the JSON shape of this action's "action" field, shown
	// to the model as a one-line example when it's deciding whether/how to
	// propose this action.
	PromptExample string

	// Validate reports whether a proposed action's arguments are well-formed
	// for this kind (e.g. a fill needs a value, a click needs a target).
	Validate func(contracts.Action) error
}

// Registry is the set of tools available to the agent loop. Iteration order
// via All matches registration order, so prompt text stays stable.
type Registry struct {
	ordered []Tool
	byKind  map[string]Tool
}

// NewRegistry builds a Registry from the given tools.
func NewRegistry(tools ...Tool) *Registry {
	r := &Registry{ordered: tools, byKind: make(map[string]Tool, len(tools))}
	for _, t := range tools {
		r.byKind[t.Kind] = t
	}
	return r
}

// Lookup returns the tool registered for kind, if any.
func (r *Registry) Lookup(kind string) (Tool, bool) {
	t, ok := r.byKind[kind]
	return t, ok
}

// All returns every registered tool, in registration order.
func (r *Registry) All() []Tool {
	return r.ordered
}
