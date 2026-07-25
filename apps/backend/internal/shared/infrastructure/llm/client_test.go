package llm

import "testing"

func TestNewSelectsProvider(t *testing.T) {
	cases := []struct {
		provider string
		want     string
	}{
		{"openai", "openaiAdapter"},
		{"google", "googleAdapter"},
		{"deepseek", "deepseekAdapter"},
		{"", "openaiAdapter"},
		{"unknown", "openaiAdapter"},
	}
	for _, tc := range cases {
		got := New(tc.provider, "http://example.com", "key", "model")
		switch got.(type) {
		case openaiAdapter:
			if tc.want != "openaiAdapter" {
				t.Errorf("provider %q: got openaiAdapter, want %s", tc.provider, tc.want)
			}
		case googleAdapter:
			if tc.want != "googleAdapter" {
				t.Errorf("provider %q: got googleAdapter, want %s", tc.provider, tc.want)
			}
		case deepseekAdapter:
			if tc.want != "deepseekAdapter" {
				t.Errorf("provider %q: got deepseekAdapter, want %s", tc.provider, tc.want)
			}
		default:
			t.Errorf("provider %q: unexpected adapter type", tc.provider)
		}
	}
}
