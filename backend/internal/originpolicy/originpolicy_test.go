package originpolicy_test

import (
	"testing"

	"github.com/clementd-tek/remote-buzzer/backend/internal/originpolicy"
)

func TestIsLocal(t *testing.T) {
	cases := map[string]bool{
		"http://localhost:5173":              true,
		"http://localhost:3000":              true,
		"http://localhost":                   true,
		"http://127.0.0.1:8080":              true,
		"http://[::1]:5173":                  true,
		"https://example.com":                false,
		"http://evil.localhost.attacker.com": false,
		"not a url at all %%%":               false,
		"":                                   false,
	}

	for origin, want := range cases {
		if got := originpolicy.IsLocal(origin); got != want {
			t.Errorf("IsLocal(%q) = %v, want %v", origin, got, want)
		}
	}
}
