package httpclient_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/elysium/elysium/cli/internal/httpclient"
)

func TestDefaultClient_NotNil(t *testing.T) {
	c := httpclient.DefaultClient()
	if c == nil {
		t.Fatal("DefaultClient() returned nil")
	}
}

func TestDefaultClient_Identity(t *testing.T) {
	// DefaultClient() must return the same pointer every time — it is a singleton.
	c1 := httpclient.DefaultClient()
	c2 := httpclient.DefaultClient()
	if c1 != c2 {
		t.Error("DefaultClient() returned different pointers on successive calls")
	}
}

func TestDefaultClient_NoTimeout(t *testing.T) {
	// DefaultClient must have no client-level timeout so context deadlines control
	// cancellation instead.
	c := httpclient.DefaultClient()
	if c.Timeout != 0 {
		t.Errorf("DefaultClient() Timeout = %v, want 0 (no timeout)", c.Timeout)
	}
}

func TestDefaultClient_HasTransport(t *testing.T) {
	c := httpclient.DefaultClient()
	if c.Transport == nil {
		t.Error("DefaultClient() Transport is nil; expected a configured *http.Transport")
	}
}

func TestClientWithTimeout_NotNil(t *testing.T) {
	c := httpclient.ClientWithTimeout(5 * time.Second)
	if c == nil {
		t.Fatal("ClientWithTimeout() returned nil")
	}
}

func TestClientWithTimeout_RespectsTimeout(t *testing.T) {
	timeout := 7 * time.Second
	c := httpclient.ClientWithTimeout(timeout)
	if c.Timeout != timeout {
		t.Errorf("ClientWithTimeout(%v) Timeout = %v, want %v", timeout, c.Timeout, timeout)
	}
}

func TestClientWithTimeout_SharesTransport(t *testing.T) {
	// All clients should share the same underlying transport for connection pooling.
	c1 := httpclient.DefaultClient()
	c2 := httpclient.ClientWithTimeout(10 * time.Second)
	if c1.Transport != c2.Transport {
		t.Error("DefaultClient and ClientWithTimeout have different transports; connection pool is NOT shared")
	}
}

func TestClientWithTimeout_IsHttpClient(t *testing.T) {
	c := httpclient.ClientWithTimeout(30 * time.Second)
	// Ensure the returned value satisfies the *http.Client type so it can be used
	// anywhere a *http.Client is expected.
	var _ *http.Client = c
}

func TestClientWithTimeout_ZeroTimeout(t *testing.T) {
	// A zero timeout should be permitted — equivalent to no timeout.
	c := httpclient.ClientWithTimeout(0)
	if c == nil {
		t.Fatal("ClientWithTimeout(0) returned nil")
	}
	if c.Timeout != 0 {
		t.Errorf("ClientWithTimeout(0) Timeout = %v, want 0", c.Timeout)
	}
}

func TestSharedTransport_MaxIdleConns(t *testing.T) {
	// Verify that the shared transport is a *http.Transport with connection-pool
	// settings present (MaxIdleConns > 0).
	c := httpclient.DefaultClient()
	transport, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("DefaultClient().Transport is %T, want *http.Transport", c.Transport)
	}
	if transport.MaxIdleConns <= 0 {
		t.Errorf("sharedTransport.MaxIdleConns = %d, want > 0", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost <= 0 {
		t.Errorf("sharedTransport.MaxIdleConnsPerHost = %d, want > 0", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout <= 0 {
		t.Errorf("sharedTransport.IdleConnTimeout = %v, want > 0", transport.IdleConnTimeout)
	}
}
