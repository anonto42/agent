package tools

import (
	"errors"
	"strings"

	"github.com/levelaxis/charli/contracts"
)

// Default returns the registry of tools Charli's agent loop currently
// supports: fill a field, click a button/link (L2), append a row to a
// connected Google Sheet (L4).
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
		Tool{
			Kind:          "sheets_append",
			Risk:          RiskConfirm,
			PromptExample: `{"kind": "sheets_append", "spreadsheetId": "<the Google Sheet ID or URL the user mentioned>", "values": ["<cell 1>", "<cell 2>", "..."]}`,
			Validate:      requireSpreadsheetAndValues,
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

func requireSpreadsheetAndValues(a contracts.Action) error {
	if strings.TrimSpace(a.SpreadsheetID) == "" {
		return errors.New("a sheets_append action needs a spreadsheetId")
	}
	if len(a.Values) == 0 {
		return errors.New("a sheets_append action needs at least one value")
	}
	return nil
}
