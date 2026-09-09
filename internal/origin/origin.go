package origin

import "strings"

// IsAllowed reports whether the request origin is permitted by the
// comma-separated allowedOrigins configuration. Use "*" to allow any origin.
//
// The allowedOrigins string is the same format consumed by the HTTP CORS
// middleware, so HTTP and WebSocket keep a single source of policy.
func IsAllowed(requestOrigin, allowedOrigins string) bool {
	if allowedOrigins == "*" {
		return true
	}
	for _, allowed := range strings.Split(allowedOrigins, ",") {
		if strings.TrimSpace(allowed) == requestOrigin {
			return true
		}
	}
	return false
}
