package validator

import (
	"fmt"
	"regexp"

	"github.com/elysium/elysium/cli/internal/emblem"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(def *emblem.Definition) []string {
	var errors []string

	if def.Name == "" {
		errors = append(errors, "name is required")
	}
	if def.Version == "" {
		errors = append(errors, "version is required")
	}
	if def.BaseURL == "" {
		errors = append(errors, "baseUrl is required")
	}

	if def.Name != "" && !isValidName(def.Name) {
		errors = append(errors, "name must be lowercase alphanumeric with dashes")
	}

	if def.Version != "" && !isValidVersion(def.Version) {
		errors = append(errors, "version must be semver (e.g., 1.0.0)")
	}

	if def.BaseURL != "" && !isValidURL(def.BaseURL) {
		errors = append(errors, "baseUrl must be a valid URL")
	}

	if len(def.Actions) == 0 {
		errors = append(errors, "at least one action is required")
	}

	for name, action := range def.Actions {
		if action.Method == "" {
			errors = append(errors, fmt.Sprintf("action '%s' missing method", name))
		}
		if action.Path == "" {
			errors = append(errors, fmt.Sprintf("action '%s' missing path", name))
		}
		if action.Method != "" && !isValidHTTPMethod(action.Method) {
			errors = append(errors, fmt.Sprintf("action '%s' has invalid method: %s", name, action.Method))
		}
	}

	return errors
}

func (v *Validator) ValidateStrict(def *emblem.Definition) []string {
	var errors []string

	for name, action := range def.Actions {
		if action.Description == "" {
			errors = append(errors, fmt.Sprintf("action '%s' should have a description", name))
		}
	}

	if def.Auth.Type == "" {
		errors = append(errors, "authentication should be defined")
	}

	return errors
}

func (v *Validator) CheckBestPractices(def *emblem.Definition) []string {
	var warnings []string

	if def.Description == "" {
		warnings = append(warnings, "consider adding a description")
	}

	return warnings
}

func isValidName(name string) bool {
	matched, _ := regexp.MatchString("^[a-z0-9-]+$", name)
	return matched
}

func isValidVersion(version string) bool {
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	return matched
}

func isValidURL(url string) bool {
	matched, _ := regexp.MatchString(`^https?://`, url)
	return matched
}

func isValidHTTPMethod(method string) bool {
	valid := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, v := range valid {
		if v == method {
			return true
		}
	}
	return false
}
