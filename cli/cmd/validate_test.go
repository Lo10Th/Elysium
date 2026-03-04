package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// minimalValidEmblem is a YAML that passes the validator with no errors.
const minimalValidEmblem = `apiVersion: v1
name: test-emblem
version: 1.0.0
description: Test emblem for validate tests
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
`

// invalidEmblem is a YAML that is syntactically parseable but fails validation
// (missing required fields).
const invalidEmblem = `apiVersion: v1
name: ""
version: ""
baseUrl: ""
actions: {}
`

// externalValidatorFailEmblem passes internal emblem validation (all fields
// non-empty) but fails the *external* validator because the name contains
// uppercase characters (nameRegex = ^[a-z0-9-]+$).
const externalValidatorFailEmblem = `apiVersion: v1
name: INVALID-UPPERCASE-NAME
version: 1.0.0
description: Test
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
`

func TestRunValidate_FileNotFound(t *testing.T) {
	err := runValidate(nil, []string{"/nonexistent/path/emblem.yaml"})
	if err == nil {
		t.Error("runValidate() expected error for missing file, got nil")
	}
}

func TestRunValidate_ValidEmblem(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(minimalValidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	err := runValidate(nil, []string{emblemPath})
	if err != nil {
		t.Errorf("runValidate() unexpected error for valid emblem: %v", err)
	}
}

func TestRunValidate_InvalidEmblem(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(invalidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	err := runValidate(nil, []string{emblemPath})
	if err == nil {
		t.Error("runValidate() expected validation error, got nil")
	}
}

// TestRunValidate_ExternalValidatorError tests the external validator path
// (passes internal emblem.Load but fails external validator format checks).
func TestRunValidate_ExternalValidatorError(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(externalValidatorFailEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	err := runValidate(nil, []string{emblemPath})
	if err == nil {
		t.Error("runValidate() expected external validation error, got nil")
	}
}

func TestRunValidate_NotYAML(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	// Writing bytes that are valid YAML syntax but fail emblem semantic
	// validation (no name, version, baseUrl) – the result must be an error,
	// not a successful validation.
	if err := os.WriteFile(emblemPath, []byte("not_a_valid_key: true\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := runValidate(nil, []string{emblemPath})
	if err == nil {
		t.Error("runValidate() expected error for non-emblem YAML, got nil")
	}
}

// minimalValidEmblemNoDesc is valid but lacks a description, triggering best-practices warnings.
const minimalValidEmblemNoDesc = `apiVersion: v1
name: test-emblem
version: 1.0.0
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
`

func TestRunValidate_Warnings(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(minimalValidEmblemNoDesc), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	// Should succeed but print best-practice warnings.
	err := runValidate(nil, []string{emblemPath})
	if err != nil {
		t.Errorf("runValidate() unexpected error for emblem with warnings: %v", err)
	}
}

// strictInvalidEmblem passes internal and external base validation but fails
// strict validation because auth.Type is not defined (ValidateStrict checks
// that authentication is explicitly configured).
const strictInvalidEmblem = `apiVersion: v1
name: test-emblem
version: 1.0.0
description: A test emblem
baseUrl: http://localhost:5000/api
actions:
  list:
    description: List items
    method: GET
    path: /items
`

func TestRunValidate_StrictMode_Fail(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(strictInvalidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	oldStrictMode := strictMode
	strictMode = true
	defer func() { strictMode = oldStrictMode }()

	err := runValidate(nil, []string{emblemPath})
	if err == nil {
		t.Error("runValidate() expected strict validation error, got nil")
	}
}

// --- runTest tests (test.go) ---

func TestRunTest_EmblemNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// No emblem.yaml in tmpDir.
	err := runTest(nil, []string{tmpDir})
	if err == nil {
		t.Error("runTest() expected error when emblem.yaml missing, got nil")
	}
}

func TestRunTest_ListActions(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(minimalValidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	// No --action flag set → should list available actions and return nil.
	oldAction := actionName
	actionName = ""
	defer func() { actionName = oldAction }()

	err := runTest(nil, []string{tmpDir})
	if err != nil {
		t.Errorf("runTest() (list actions) unexpected error: %v", err)
	}
}

func TestRunTest_InvalidAction(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(minimalValidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	oldAction := actionName
	actionName = "nonexistent-action"
	defer func() { actionName = oldAction }()

	err := runTest(nil, []string{tmpDir})
	if err == nil {
		t.Error("runTest() expected error for invalid action, got nil")
	}
}

func TestRunTest_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	emblemPath := filepath.Join(tmpDir, "emblem.yaml")

	if err := os.WriteFile(emblemPath, []byte(minimalValidEmblem), 0644); err != nil {
		t.Fatalf("failed to write test emblem: %v", err)
	}

	oldAction := actionName
	oldDryRun := dryRunFlag
	actionName = "list"
	dryRunFlag = true
	defer func() {
		actionName = oldAction
		dryRunFlag = oldDryRun
	}()

	err := runTest(nil, []string{tmpDir})
	if err != nil {
		t.Errorf("runTest() (dry-run) unexpected error: %v", err)
	}
}
