// Package originpolicy centralizes the "is this origin allowed to talk to
// us" decision so the REST CORS middleware and the websocket upgrader
// can't drift out of sync with each other.
package originpolicy

import (
	"net"
	"net/url"
)

// IsLocal reports whether origin is some flavor of localhost, regardless
// of port. Dev servers (Vite included) silently pick a different port
// than their default whenever the default is already taken, so pinning a
// single hardcoded "http://localhost:5173" as the only allowed origin is
// a common, confusing trap: REST calls keep working (same-origin through
// the dev proxy) while anything origin-checked directly — namely the
// websocket upgrade — silently starts failing.
func IsLocal(origin string) bool {
	host := hostname(origin)
	if host == "" {
		return false
	}

	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// IsPrivateNetwork reports whether origin is served from a private
// (RFC-1918 / RFC-4193) IP address — i.e. a LAN address like 192.168.x.x,
// 10.x.x.x, or 172.16–31.x.x.
//
// This allows any device on the same local network to reach the app
// without having to list every possible LAN IP in FRONTEND_ORIGINS.
// It does NOT open the app to the public internet.
func IsPrivateNetwork(origin string) bool {
	host := hostname(origin)
	if host == "" {
		return false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7", // IPv6 ULA
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// IsAllowed is the single entry-point used by both the CORS middleware
// and the WebSocket upgrader. An origin is allowed when:
//
//  1. It is localhost / 127.0.0.1 / ::1 (always OK for local dev).
//  2. It is a private LAN address (always OK for local-network testing).
//  3. It is explicitly listed in the configured allowed-origins set.
//  4. No explicit origins are configured (open / trust-all mode).
func IsAllowed(origin string, allowed map[string]bool) bool {
	if IsLocal(origin) || IsPrivateNetwork(origin) {
		return true
	}

	if len(allowed) == 0 {
		return true
	}

	return allowed[origin]
}

func hostname(origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}

	return u.Hostname()
}
