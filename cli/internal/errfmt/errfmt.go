package errfmt

import (
	"fmt"
	"strings"
)

type DetailedError struct {
	Err        error
	Reason     string
	Suggestion string
	Context    map[string]string
}

func (e *DetailedError) Error() string {
	var msg strings.Builder

	msg.WriteString(colorize("Error: ", colorRed))
	msg.WriteString(colorize(e.Err.Error(), colorRed))
	msg.WriteString("\n")

	if e.Reason != "" {
		msg.WriteString(colorize("  Reason: ", colorGray))
		msg.WriteString(e.Reason)
		msg.WriteString("\n")
	}

	for key, value := range e.Context {
		msg.WriteString(fmt.Sprintf("  %s: %s\n", colorize(key, colorGray), value))
	}

	if e.Suggestion != "" {
		msg.WriteString(colorize("  Suggestion: ", colorCyan))
		msg.WriteString(e.Suggestion)
		msg.WriteString("\n")
	}

	return msg.String()
}

func NewDetailedError(err error) *DetailedError {
	return &DetailedError{Err: err}
}

func (e *DetailedError) WithReason(reason string) *DetailedError {
	e.Reason = reason
	return e
}

func (e *DetailedError) WithSuggestion(suggestion string) *DetailedError {
	e.Suggestion = suggestion
	return e
}

func (e *DetailedError) WithContext(key, value string) *DetailedError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

func ConnectionError(url string, err error) error {
	return NewDetailedError(err).
		WithReason("Connection refused").
		WithContext("URL", url).
		WithSuggestion("Is the API server running?\n               Try: Check if the server is started")
}

func AuthRequiredError(keyName string) error {
	return NewDetailedError(fmt.Errorf("authentication required")).
		WithReason(fmt.Sprintf("Missing API key: %s", keyName)).
		WithSuggestion(fmt.Sprintf("Set your API key:\n               export %s=your-key-here\n               Or: ely keys create --name my-key", keyName))
}

func EmblemNotFoundError(name string) error {
	return NewDetailedError(fmt.Errorf("emblem '%s' not found", name)).
		WithSuggestion(fmt.Sprintf("Check available emblems:\n               ely search %s\n               Or: ely pull %s", name, name))
}

func InvalidYAMLError(filePath string, err error) error {
	return NewDetailedError(err).
		WithContext("File", filePath).
		WithSuggestion("Check YAML syntax at https://www.yamllint.com/")
}

func RateLimitError(retryAfter int) error {
	return NewDetailedError(fmt.Errorf("rate limit exceeded")).
		WithReason("Too many requests").
		WithContext("Retry after", fmt.Sprintf("%d seconds", retryAfter)).
		WithSuggestion("Wait before retrying or upgrade your plan")
}

func NetworkError(err error) error {
	return NewDetailedError(err).
		WithReason("Network connectivity issue").
		WithSuggestion("Check your internet connection and try again")
}

func ConfigNotFoundError() error {
	return NewDetailedError(fmt.Errorf("configuration not found")).
		WithReason("No config file exists").
		WithSuggestion("Run: ely login to authenticate\n               Or create ~/.elysium/config.yaml manually")
}

func PermissionError(resource string) error {
	return NewDetailedError(fmt.Errorf("permission denied")).
		WithReason(fmt.Sprintf("Insufficient permissions for %s", resource)).
		WithSuggestion("Check your API key permissions or contact support")
}

func APIError(statusCode int, message string) error {
	err := fmt.Errorf("API returned status %d: %s", statusCode, message)

	switch statusCode {
	case 401:
		return AuthRequiredError("API_KEY")
	case 403:
		return PermissionError("this resource")
	case 404:
		return NewDetailedError(fmt.Errorf("resource not found")).
			WithReason(message)
	case 429:
		return RateLimitError(60)
	case 500:
		return NewDetailedError(err).
			WithReason("API server error").
			WithSuggestion("Try again later or contact support")
	default:
		return NewDetailedError(err).
			WithContext("Status", fmt.Sprintf("%d", statusCode))
	}
}
