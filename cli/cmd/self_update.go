package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/elysium/elysium/cli/internal/selfupdate"
	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update ely CLI to the latest version",
	Long: `Check GitHub for the latest release and update the ely binary in place.

Examples:
  # Update to the latest version
  ely self-update

  # Check for updates without installing
  ely self-update --check

  # Install a specific version
  ely self-update --version v0.3.0

  # Force update even if already on the latest version
  ely self-update --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")
		targetVersion, _ := cmd.Flags().GetString("version")
		force, _ := cmd.Flags().GetBool("force")

		if checkOnly {
			return checkForUpdates()
		}
		return performSelfUpdate(targetVersion, force)
	},
}

// checkForUpdates fetches the latest release and reports whether an update is available.
func checkForUpdates() error {
	fmt.Println("Checking for updates...")

	release, err := selfupdate.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("could not fetch release info: %w", err)
	}

	current := Version
	latest := release.TagName

	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Latest version:  %s\n", latest)

	if !selfupdate.IsNewer(current, latest) {
		fmt.Println("✓ Already up to date.")
		return nil
	}

	fmt.Printf("Update available: run 'ely self-update' to install %s.\n", latest)
	return nil
}

// performSelfUpdate downloads and installs the specified (or latest) version.
func performSelfUpdate(targetVersion string, force bool) error {
	fmt.Println("Checking for updates...")

	var release *selfupdate.Release
	var err error

	if targetVersion != "" {
		tag := targetVersion
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		release, err = selfupdate.GetReleaseByTag(tag)
	} else {
		release, err = selfupdate.GetLatestRelease()
	}
	if err != nil {
		return fmt.Errorf("could not fetch release info: %w", err)
	}

	current := Version
	latest := release.TagName

	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Latest version:  %s\n", latest)

	if !force && !selfupdate.IsNewer(current, latest) {
		fmt.Println("✓ Already up to date. Use --force to reinstall.")
		return nil
	}

	downloadURL, err := selfupdate.FindAssetURL(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading ely %s...\n", latest)
	tmpPath, err := selfupdate.DownloadBinary(downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	if err := selfupdate.ReplaceBinary(tmpPath); err != nil {
		return err
	}

	fmt.Printf("✓ Updated successfully to %s\n", selfupdate.NormalizeVersion(latest))
	return nil
}

func init() {
	selfUpdateCmd.Flags().Bool("check", false, "check for updates without installing")
	selfUpdateCmd.Flags().String("version", "", "install a specific version (e.g. v0.3.0)")
	selfUpdateCmd.Flags().Bool("force", false, "reinstall even if already on the latest version")
	rootCmd.AddCommand(selfUpdateCmd)
}
