package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ReplaceBinary atomically replaces the current executable with the binary at newBinaryPath.
// On Windows, the running executable cannot be replaced directly, so a "pending rename" approach
// is used instead: the old binary is renamed to ely-old.exe and the new one to ely.exe.
func ReplaceBinary(newBinaryPath string) error {
	// Make the new binary executable.
	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable path: %w", err)
	}

	if runtime.GOOS == "windows" {
		return replaceWindows(execPath, newBinaryPath)
	}

	// Unix: atomic rename. The old file is replaced in-place.
	if err := os.Rename(newBinaryPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary (try running with elevated privileges): %w", err)
	}

	return nil
}

// replaceWindows handles binary replacement on Windows.
// Windows does not allow renaming a running executable, so we:
//  1. Rename the current binary to <name>-old<ext>
//  2. Move the new binary to the current binary path
//  3. Instruct the user to delete the old binary after restarting.
func replaceWindows(execPath, newBinaryPath string) error {
	ext := filepath.Ext(execPath)
	base := execPath[:len(execPath)-len(ext)]
	oldPath := base + "-old" + ext

	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary: %w", err)
	}

	if err := os.Rename(newBinaryPath, execPath); err != nil {
		// Try to restore the original binary.
		_ = os.Rename(oldPath, execPath)
		return fmt.Errorf("failed to move new binary into place: %w", err)
	}

	fmt.Printf("ℹ️  Old binary kept as %s — delete it after restarting your terminal.\n", oldPath)
	return nil
}
