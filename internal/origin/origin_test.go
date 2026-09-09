package origin

import "testing"

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		name           string
		requestOrigin  string
		allowedOrigins string
		want           bool
	}{
		{name: "wildcard allows anything", requestOrigin: "https://example.com", allowedOrigins: "*", want: true},
		{name: "exact match", requestOrigin: "https://example.com", allowedOrigins: "https://example.com", want: true},
		{name: "match in list with whitespace", requestOrigin: "https://two.com", allowedOrigins: "https://one.com, https://two.com", want: true},
		{name: "no match", requestOrigin: "https://evil.com", allowedOrigins: "https://example.com", want: false},
		{name: "empty allowed list denies", requestOrigin: "https://example.com", allowedOrigins: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllowed(tt.requestOrigin, tt.allowedOrigins); got != tt.want {
				t.Errorf("IsAllowed(%q, %q) = %v; want %v", tt.requestOrigin, tt.allowedOrigins, got, tt.want)
			}
		})
	}
}
