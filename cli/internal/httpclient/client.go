// Package httpclient provides a shared, connection-pooling HTTP client for the
// Elysium CLI. Using a single transport across all HTTP calls avoids the cost of
// opening a new TCP connection (and TLS handshake) on every request.
package httpclient

import (
	"net/http"
	"time"
)

// sharedTransport is the single transport shared by every client returned from
// this package.  It is configured for efficient reuse:
//   - up to 100 idle connections total
//   - up to 10 idle connections per host
//   - idle connections kept alive for 90 s
//   - TLS handshake capped at 10 s
var sharedTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
}

// defaultClient is the process-wide shared HTTP client.  It has no client-level
// timeout so that callers can use per-request context deadlines.
var defaultClient = &http.Client{
	Transport: sharedTransport,
}

// DefaultClient returns the process-wide shared HTTP client.
// The client has no client-level timeout; use context.WithTimeout together with
// http.NewRequestWithContext to set per-request deadlines.
func DefaultClient() *http.Client {
	return defaultClient
}

// ClientWithTimeout returns a new *http.Client that shares the connection pool
// with DefaultClient() but applies timeout as a client-level deadline on every
// request.
//
// The returned client is intended to be stored in a package-level variable and
// reused across requests, not created anew for each call.
func ClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: sharedTransport,
		Timeout:   timeout,
	}
}
