// Package originpolicy centralizes the "is this origin allowed to talk to
// us" decision so the REST CORS middleware and the websocket upgrader
// can't drift out of sync with each other.
package originpolicy

import "net/url"

// IsLocal reports whether origin is some flavor of localhost, regardless
// of port. Dev servers (Vite included) silently pick a different port
// than their default whenever the default is already taken, so pinning a
// single hardcoded "http://localhost:5173" as the only allowed origin is
// a common, confusing trap: REST calls keep working (same-origin through
// the dev proxy) while anything origin-checked directly — namely the
// websocket upgrade — silently starts failing.
func IsLocal(origin string) bool {
	u, err := url.Parse(origin)

	if err != nil {
		return false
	}

	host := u.Hostname()

	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
