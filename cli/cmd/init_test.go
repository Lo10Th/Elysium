package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_Success(t *testing.T) {
	tmpDir := t.TempDir()

	oldOutput := outputFlag
	oldCategory := categoryFlag
	oldDescription := descriptionFlag
	outputFlag = filepath.Join(tmpDir, "my-api")
	categoryFlag = "general"
	descriptionFlag = "A test API"
	defer func() {
		outputFlag = oldOutput
		categoryFlag = oldCategory
		descriptionFlag = oldDescription
	}()

	err := runInit(nil, []string{"my-api"})
	if err != nil {
		t.Errorf("runInit() unexpected error: %v", err)
	}

	// Verify that emblem.yaml was created.
	emblemPath := filepath.Join(tmpDir, "my-api", "emblem.yaml")
	if _, err := os.Stat(emblemPath); os.IsNotExist(err) {
		t.Errorf("runInit() emblem.yaml not created at %s", emblemPath)
	}
}

func TestRunInit_InvalidName(t *testing.T) {
	// Names with spaces should fail scaffold.ValidateName.
	err := runInit(nil, []string{"invalid name with spaces"})
	if err == nil {
		t.Error("runInit() expected error for invalid name, got nil")
	}
}

func TestRunInit_DirectoryAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "already-exists")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}

	// Reset outputFlag to "." so runInit constructs the path from the name arg.
	oldOutput := outputFlag
	outputFlag = "."
	defer func() { outputFlag = oldOutput }()

	// Change working directory to tmpDir so the relative path resolves correctly.
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(oldWD) //nolint:errcheck

	err = runInit(nil, []string{"already-exists"})
	if err == nil {
		t.Error("runInit() expected error for existing directory, got nil")
	}
}

func TestRunInit_PaymentsCategory(t *testing.T) {
	tmpDir := t.TempDir()

	oldOutput := outputFlag
	oldCategory := categoryFlag
	outputFlag = filepath.Join(tmpDir, "payment-api")
	categoryFlag = "payments"
	defer func() {
		outputFlag = oldOutput
		categoryFlag = oldCategory
	}()

	err := runInit(nil, []string{"payment-api"})
	if err != nil {
		t.Errorf("runInit() (payments category) unexpected error: %v", err)
	}

	emblemPath := filepath.Join(tmpDir, "payment-api", "emblem.yaml")
	if _, statErr := os.Stat(emblemPath); os.IsNotExist(statErr) {
		t.Errorf("runInit() emblem.yaml not created at %s", emblemPath)
	}
}

func TestRunInit_EcommerceCategory(t *testing.T) {
	tmpDir := t.TempDir()

	oldOutput := outputFlag
	oldCategory := categoryFlag
	outputFlag = filepath.Join(tmpDir, "shop-api")
	categoryFlag = "ecommerce"
	defer func() {
		outputFlag = oldOutput
		categoryFlag = oldCategory
	}()

	err := runInit(nil, []string{"shop-api"})
	if err != nil {
		t.Errorf("runInit() (ecommerce category) unexpected error: %v", err)
	}
}
