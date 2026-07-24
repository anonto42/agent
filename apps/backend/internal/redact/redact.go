// Package redact scrubs sensitive text out of page content before it's sent
// to the LLM. See .agents/shared/rules/agent-safety.md: "Sensitive fields
// (passwords, card numbers, tokens) are REDACTED before any page content is
// sent to the LLM." This is deliberately narrow — only the categories that
// rule names — so it doesn't degrade L1's ability to answer ordinary
// questions about a page (e.g. a visible contact email must stay readable).
package redact

import "regexp"

const placeholder = "[REDACTED]"

var (
	ssnPattern      = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	passwordPattern = regexp.MustCompile(`(?i)password\s*[:=]\s*\S+`)
	// Token/API-key prefixes from major providers, plus a generic "Bearer <token>".
	tokenPattern = regexp.MustCompile(`\b(?:sk|ghp|gho|ghu|ghs|ghr|pk)[-_][A-Za-z0-9]{16,}\b|\bAKIA[0-9A-Z]{16}\b|\bBearer\s+[A-Za-z0-9\-_.]+`)
	// Candidate card numbers: 13-19 digits, optionally grouped by spaces or
	// dashes. Confirmed with a Luhn check before redacting, so ordinary
	// digit runs (order numbers, tracking numbers, ids) are left alone.
	cardCandidatePattern = regexp.MustCompile(`\b(?:\d[ -]?){12,18}\d\b`)
)

// Text returns s with passwords, card numbers, and API tokens/keys replaced
// with a placeholder.
func Text(s string) string {
	s = passwordPattern.ReplaceAllString(s, placeholder)
	s = tokenPattern.ReplaceAllString(s, placeholder)
	s = ssnPattern.ReplaceAllString(s, placeholder)
	s = cardCandidatePattern.ReplaceAllStringFunc(s, func(match string) string {
		if isLuhnValid(match) {
			return placeholder
		}
		return match
	})
	return s
}

// isLuhnValid reports whether the digits in s (ignoring spaces/dashes) pass
// the Luhn checksum used by card numbers.
func isLuhnValid(s string) bool {
	var digits []int
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits = append(digits, int(r-'0'))
		case r == ' ' || r == '-':
			continue
		default:
			return false
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
