package redact

import (
	"strings"
	"testing"
)

func TestTextRedactsCardNumbers(t *testing.T) {
	for _, s := range []string{
		"Card on file: 4111 1111 1111 1111",
		"Card on file: 4111-1111-1111-1111",
		"Card on file: 4111111111111111",
	} {
		got := Text(s)
		if strings.Contains(got, "1111") {
			t.Errorf("expected card number redacted in %q, got %q", s, got)
		}
		if !strings.Contains(got, "[REDACTED]") {
			t.Errorf("expected a redaction placeholder in %q, got %q", s, got)
		}
	}
}

func TestTextLeavesOrdinaryDigitRunsAlone(t *testing.T) {
	// A tracking/order number that happens to be long but fails the Luhn
	// check must not be treated as a card number.
	s := "Order number: 1234567890123"
	got := Text(s)
	if got != s {
		t.Errorf("expected non-card digit run to be left alone, got %q", got)
	}
}

func TestTextRedactsSSN(t *testing.T) {
	got := Text("SSN: 123-45-6789 on file")
	if strings.Contains(got, "123-45-6789") || !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected SSN redacted, got %q", got)
	}
}

func TestTextRedactsPassword(t *testing.T) {
	for _, s := range []string{"password: hunter2", "Password=hunter2"} {
		got := Text(s)
		if strings.Contains(got, "hunter2") {
			t.Errorf("expected password value redacted in %q, got %q", s, got)
		}
	}
}

func TestTextRedactsTokens(t *testing.T) {
	for _, s := range []string{
		"key is sk-abcdefghijklmnopqrstuvwxyz",
		"token ghp_abcdefghijklmnopqrstuvwx",
		"AWS key AKIAABCDEFGHIJKLMNOP",
		"Authorization: Bearer abc123.def456.ghi789",
	} {
		got := Text(s)
		if !strings.Contains(got, "[REDACTED]") {
			t.Errorf("expected a token redacted in %q, got %q", s, got)
		}
	}
}

func TestTextLeavesOrdinaryContentAlone(t *testing.T) {
	s := "Contact us at hello@example.com or call our office. This page has 42 items in stock."
	got := Text(s)
	if got != s {
		t.Errorf("expected ordinary page text untouched, got %q", got)
	}
}
