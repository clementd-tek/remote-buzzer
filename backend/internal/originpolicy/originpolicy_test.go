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
		"http://127.0.0.1:8080":             true,
		"http://[::1]:5173":                 true,
		"https://example.com":               false,
		"http://evil.localhost.attacker.com": false,
		"not a url at all %%%":              false,
		"":                                  false,
	}

	for origin, want := range cases {
		if got := originpolicy.IsLocal(origin); got != want {
			t.Errorf("IsLocal(%q) = %v, want %v", origin, got, want)
		}
	}
}

func TestIsPrivateNetwork(t *testing.T) {
	cases := map[string]bool{
		// RFC-1918 ranges
		"http://192.168.1.42:8080": true,
		"http://192.168.0.1":       true,
		"http://10.0.0.1:3000":     true,
		"http://10.255.255.255":    true,
		"http://172.16.0.1":        true,
		"http://172.31.255.255":    true,
		// IPv6 ULA
		"http://[fc00::1]:8080": true,
		"http://[fd12:3456::1]": true,
		// Public addresses — must not be allowed by this check
		"https://example.com":      false,
		"http://8.8.8.8":           false,
		"http://172.32.0.1":        false, // just outside 172.16/12
		"http://localhost:5173":    false, // local, not private-network range
		"http://127.0.0.1":         false, // loopback, not private-network range
		"not a url at all %%%":     false,
		"":                         false,
	}

	for origin, want := range cases {
		if got := originpolicy.IsPrivateNetwork(origin); got != want {
			t.Errorf("IsPrivateNetwork(%q) = %v, want %v", origin, got, want)
		}
	}
}

func TestIsAllowed(t *testing.T) {
	configured := map[string]bool{
		"https://buzzer.example.com": true,
	}
	empty := map[string]bool{}

	cases := []struct {
		origin  string
		allowed map[string]bool
		want    bool
	}{
		// localhost always allowed regardless of configured list
		{"http://localhost:5173", configured, true},
		{"http://127.0.0.1:8080", configured, true},
		// LAN IPs always allowed
		{"http://192.168.1.42:8080", configured, true},
		{"http://10.0.0.5:8080", configured, true},
		// Explicitly listed origin allowed
		{"https://buzzer.example.com", configured, true},
		// Unlisted public origin blocked
		{"https://evil.com", configured, false},
		// Empty allowlist = trust-all (open mode)
		{"https://anything.example.com", empty, true},
	}

	for _, tc := range cases {
		if got := originpolicy.IsAllowed(tc.origin, tc.allowed); got != tc.want {
			t.Errorf("IsAllowed(%q, ...) = %v, want %v", tc.origin, got, tc.want)
		}
	}
}
