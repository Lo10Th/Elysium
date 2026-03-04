package cmd

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateRandomState(t *testing.T) {
	state1, err := generateRandomState()
	if err != nil {
		t.Fatalf("generateRandomState() returned error: %v", err)
	}

	if state1 == "" {
		t.Error("generateRandomState() returned empty string")
	}

	state2, err := generateRandomState()
	if err != nil {
		t.Fatalf("generateRandomState() returned error: %v", err)
	}

	if state1 == state2 {
		t.Error("generateRandomState() returned same state twice")
	}

	decoded, err := base64.URLEncoding.DecodeString(state1)
	if err != nil {
		t.Fatalf("state is not valid base64: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("state decoded to wrong length: got %d, want 32", len(decoded))
	}
}

func TestIsCommandAvailable(t *testing.T) {
	lsAvailable := isCommandAvailable("ls")
	if !lsAvailable {
		t.Error("isCommandAvailable('ls') returned false, expected true")
	}

	fakeAvailable := isCommandAvailable("this-command-does-not-exist-12345")
	if fakeAvailable {
		t.Error("isCommandAvailable('this-command-does-not-exist-12345') returned true, expected false")
	}
}

func TestOpenBrowser(t *testing.T) {
	err := openBrowser("https://example.com")
	if err != nil {
		if strings.Contains(err.Error(), "no browser command available") {
			return
		}
	}
}

func BenchmarkGenerateRandomState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generateRandomState()
		if err != nil {
			b.Fatalf("generateRandomState returned error: %v", err)
		}
	}
}
