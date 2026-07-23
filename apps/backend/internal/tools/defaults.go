package tools

import (
	"errors"
	"strings"

	"github.com/levelaxis/charli/contracts"
)

// Default returns the registry of tools Charli's L2 agent loop currently
// supports: fill a field, click a button/link.
func Default() *Registry {
	return NewRegistry(
		Tool{
			Kind:          "fill",
			Risk:          RiskAuto,
			PromptExample: `{"kind": "fill", "target": "<field description>", "value": "<text to enter>"}`,
			Validate:      requireValue,
		},
		Tool{
			Kind:          "click",
			Risk:          RiskConfirm,
			PromptExample: `{"kind": "click", "target": "<button/link text>"}`,
			Validate:      requireTarget,
		},
	)
}

func requireValue(a contracts.Action) error {
	if strings.TrimSpace(a.Value) == "" {
		return errors.New("a fill action needs a value")
	}
	return nil
}

func requireTarget(a contracts.Action) error {
	if strings.TrimSpace(a.Target) == "" {
		return errors.New("a click action needs a target")
	}
	return nil
}
