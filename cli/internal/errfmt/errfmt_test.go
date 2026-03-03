package errfmt

import (
	"errors"
	"strings"
	"testing"
)

func TestConnectionError(t *testing.T) {
	err := ConnectionError("http://localhost:5000", errors.New("dial tcp: connection refused"))

	msg := err.Error()
	if !strings.Contains(msg, "Connection refused") {
		t.Error("Should contain reason")
	}
	if !strings.Contains(msg, "localhost:5000") {
		t.Error("Should contain URL")
	}
	if !strings.Contains(msg, "Suggestion") {
		t.Error("Should contain suggestion")
	}
}

func TestAuthRequiredError(t *testing.T) {
	err := AuthRequiredError("CLOTHING_SHOP_API_KEY")

	msg := err.Error()
	if !strings.Contains(msg, "authentication required") {
		t.Error("Should contain error message")
	}
	if !strings.Contains(msg, "export CLOTHING_SHOP_API_KEY") {
		t.Error("Should contain export command")
	}
	if !strings.Contains(msg, "ely keys create") {
		t.Error("Should contain keys create command")
	}
}

func TestEmblemNotFoundError(t *testing.T) {
	err := EmblemNotFoundError("my-api")

	msg := err.Error()
	if !strings.Contains(msg, "my-api") {
		t.Error("Should contain emblem name")
	}
	if !strings.Contains(msg, "not found") {
		t.Error("Should contain 'not found'")
	}
	if !strings.Contains(msg, "ely search") {
		t.Error("Should contain search suggestion")
	}
}

func TestInvalidYAMLError(t *testing.T) {
	err := InvalidYAMLError("emblem.yaml", errors.New("yaml: line 15: unexpected mapping key"))

	msg := err.Error()
	if !strings.Contains(msg, "emblem.yaml") {
		t.Error("Should contain file path")
	}
	if !strings.Contains(msg, "line 15") {
		t.Error("Should contain line number from error")
	}
	if !strings.Contains(msg, "yamllint.com") {
		t.Error("Should contain yamllint URL")
	}
}

func TestRateLimitError(t *testing.T) {
	err := RateLimitError(60)

	msg := err.Error()
	if !strings.Contains(msg, "rate limit exceeded") {
		t.Error("Should contain rate limit message")
	}
	if !strings.Contains(msg, "60 seconds") {
		t.Error("Should contain retry after time")
	}
	if !strings.Contains(msg, "upgrade your plan") {
		t.Error("Should contain upgrade suggestion")
	}
}

func TestNetworkError(t *testing.T) {
	err := NetworkError(errors.New("network is unreachable"))

	msg := err.Error()
	if !strings.Contains(msg, "Network connectivity issue") {
		t.Error("Should contain network reason")
	}
	if !strings.Contains(msg, "internet connection") {
		t.Error("Should contain internet connection suggestion")
	}
}

func TestConfigNotFoundError(t *testing.T) {
	err := ConfigNotFoundError()

	msg := err.Error()
	if !strings.Contains(msg, "configuration not found") {
		t.Error("Should contain config not found message")
	}
	if !strings.Contains(msg, "ely login") {
		t.Error("Should contain login command")
	}
	if !strings.Contains(msg, ".elysium/config.yaml") {
		t.Error("Should contain config path")
	}
}

func TestPermissionError(t *testing.T) {
	err := PermissionError("this resource")

	msg := err.Error()
	if !strings.Contains(msg, "permission denied") {
		t.Error("Should contain permission denied message")
	}
	if !strings.Contains(msg, "this resource") {
		t.Error("Should contain resource name")
	}
	if !strings.Contains(msg, "API key permissions") {
		t.Error("Should contain API key suggestion")
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		checks     []string
	}{
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			message:    "Unauthorized",
			checks:     []string{"authentication required", "API_KEY"},
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			message:    "Forbidden",
			checks:     []string{"permission denied", "Insufficient permissions"},
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			message:    "Resource not found",
			checks:     []string{"resource not found", "Resource not found"},
		},
		{
			name:       "429 Rate Limit",
			statusCode: 429,
			message:    "Too many requests",
			checks:     []string{"rate limit exceeded", "60 seconds"},
		},
		{
			name:       "500 Server Error",
			statusCode: 500,
			message:    "Internal server error",
			checks:     []string{"API returned status 500", "server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := APIError(tt.statusCode, tt.message)
			msg := err.Error()
			for _, check := range tt.checks {
				if !strings.Contains(msg, check) {
					t.Errorf("Expected '%s' in error message for %s", check, tt.name)
				}
			}
		})
	}
}

func TestDetailedErrorWithContext(t *testing.T) {
	err := NewDetailedError(errors.New("test error")).
		WithReason("This is the reason").
		WithContext("Key", "Value").
		WithContext("Another", "Entry").
		WithSuggestion("Try this solution")

	msg := err.Error()
	if !strings.Contains(msg, "test error") {
		t.Error("Should contain error message")
	}
	if !strings.Contains(msg, "This is the reason") {
		t.Error("Should contain reason")
	}
	if !strings.Contains(msg, "Key") || !strings.Contains(msg, "Value") {
		t.Error("Should contain first context entry")
	}
	if !strings.Contains(msg, "Another") || !strings.Contains(msg, "Entry") {
		t.Error("Should contain second context entry")
	}
	if !strings.Contains(msg, "Try this solution") {
		t.Error("Should contain suggestion")
	}
}

func TestColorDisabled(t *testing.T) {
	original := colorEnabled
	t.Run("colors enabled", func(t *testing.T) {
		colorEnabled = true
		msg := colorize("test", colorRed)
		if !strings.Contains(msg, "\033[31m") {
			t.Error("Should contain ANSI color code when enabled")
		}
	})

	t.Run("colors disabled", func(t *testing.T) {
		colorEnabled = false
		msg := colorize("test", colorRed)
		if strings.Contains(msg, "\033[") {
			t.Error("Should not contain ANSI codes when disabled")
		}
	})
	colorEnabled = original
}
