// Package domain models Charli's audit trail: one Entry per tool call
// outcome, per .agents/shared/rules/agent-safety.md — "Every tool call
// (user, tool, args, result, timestamp) is written to the audit log. No
// silent actions."
package domain

import "time"

// Outcome values an Entry can carry. Kept as plain strings (not a Go enum)
// since this is what a future audit-log viewer would filter/display on
// directly.
const (
	OutcomeProposed     = "proposed"      // shown to the user for confirmation
	OutcomeAutoExecuted = "auto_executed" // RiskAuto tool, no confirmation needed
	OutcomeDenied       = "denied"        // safety engine rejected it
	OutcomeRejected     = "rejected"      // user rejected a proposed action
	OutcomeExecuted     = "executed"      // user approved; told to run
	OutcomeObservedOK   = "observed_ok"   // the executed action succeeded on the page
	OutcomeObservedFail = "observed_failed"
	OutcomeInterrupted  = "interrupted" // user's kill switch stopped the task
)

// Entry is one audit-logged event. Deliberately not persisted via
// gorm.Model: no UpdatedAt/DeletedAt — an audit trail is append-only, never
// edited or soft-deleted.
type Entry struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Session   string    `gorm:"index"`
	TaskID    string
	Kind      string // action kind, e.g. "fill" | "click"; empty when not action-specific
	Target    string
	Outcome   string `gorm:"index"`
	Reason    string // set for denials and failed observations
}

// Repository persists audit entries. Defined here so application code
// depends on this interface, not a concrete GORM type.
type Repository interface {
	Record(Entry) error
}
