package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	checkUpdatesNameColumnWidth     = 24
	checkUpdatesCurrentColumnWidth  = 12
	checkUpdatesLatestColumnWidth   = 12
)

// securityIcon returns a display string for a security severity level.
func securityIcon(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴 CRITICAL"
	case "high":
		return "🔴 HIGH"
	case "medium":
		return "⚠️  MEDIUM"
	case "low":
		return "⚠️  LOW"
	default:
		return "✓"
	}
}

var checkUpdatesCmd = &cobra.Command{
	Use:   "check-updates",
	Short: "Check for updates and security advisories for installed emblems",
	Long: `Query the registry for the latest version of each installed emblem
and display a table showing available updates and any security advisories.

Examples:
  # Check all installed emblems for updates
  ely check-updates

  # Disable update notifications globally
  ely check-updates --no-check`,
	RunE: func(cmd *cobra.Command, args []string) error {
		noCheck, _ := cmd.Flags().GetBool("no-check")
		if noCheck {
			fmt.Println("Update checks are disabled (--no-check).")
			return nil
		}

		cfg := config.Get()

		if len(cfg.Installed) == 0 {
			fmt.Println("No emblems installed.")
			return nil
		}

		client := api.NewClient()
		client.SetToken(cfg.Token)

		type updateEntry struct {
			name             string
			current          string
			latest           string
			securityAdvisory string
			securitySeverity string
		}

		var entries []updateEntry
		var errors []string

		for name, currentVersion := range cfg.Installed {
			emblemInfo, err := client.GetEmblem(name)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					return err
				}
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
				continue
			}

			// Refresh the version cache.
			if cacheErr := config.SetVersionCache(
				name,
				emblemInfo.LatestVersion,
				emblemInfo.SecurityAdvisory,
				emblemInfo.SecuritySeverity,
			); cacheErr != nil {
				// Non-fatal cache error; continue.
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not cache version info for %s: %v\n", name, cacheErr)
			}

			entries = append(entries, updateEntry{
				name:             name,
				current:          currentVersion,
				latest:           emblemInfo.LatestVersion,
				securityAdvisory: emblemInfo.SecurityAdvisory,
				securitySeverity: emblemInfo.SecuritySeverity,
			})
		}

		// Record the time of this check.
		_ = config.SetLastUpdateCheck()

		for _, e := range errors {
			fmt.Printf("Warning: %s\n", e)
		}

		if len(entries) == 0 {
			fmt.Println("All installed emblems are up to date.")
			return nil
		}

		sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

		// Print header.
		fmt.Printf("%-*s %-*s %-*s %s\n",
			checkUpdatesNameColumnWidth, "NAME",
			checkUpdatesCurrentColumnWidth, "CURRENT",
			checkUpdatesLatestColumnWidth, "LATEST",
			"SECURITY",
		)

		anyOutdated := false
		for _, e := range entries {
			security := securityIcon(e.securitySeverity)
			if e.securityAdvisory != "" && e.securitySeverity == "" {
				security = "⚠️  ADVISORY"
			}

			fmt.Printf("%-*s %-*s %-*s %s\n",
				checkUpdatesNameColumnWidth, e.name,
				checkUpdatesCurrentColumnWidth, e.current,
				checkUpdatesLatestColumnWidth, e.latest,
				security,
			)

			if e.current != e.latest {
				anyOutdated = true
			}

			if e.securityAdvisory != "" {
				fmt.Printf("  🔴 SECURITY UPDATE: %s %s has vulnerability %s\n",
					e.name, e.current, e.securityAdvisory)
				fmt.Printf("     Update to %s immediately!\n", e.latest)
			}
		}

		if anyOutdated {
			fmt.Println()
			fmt.Println("Run 'ely update <name>' to update an emblem, or 'ely update --all' to update all.")
		}

		return nil
	},
}

// PrintUpdateNotification shows a brief notification if the emblem has a cached update available.
// It is called before executing an emblem action. It is a no-op when noCheck is true.
func PrintUpdateNotification(emblemName string, noCheck bool) {
	if noCheck {
		return
	}

	currentVersion, installed := config.GetInstalledVersion(emblemName)
	if !installed {
		return
	}

	cached, ok := config.GetVersionCache(emblemName)
	if !ok {
		// No cached data yet; suggest a check if last full check is over 24 hours ago.
		last := config.GetLastUpdateCheck()
		if last.IsZero() || time.Since(last) > 24*time.Hour {
			fmt.Printf("ℹ️  Run 'ely check-updates' to check for updates.\n")
		}
		return
	}

	if cached.SecurityAdvisory != "" {
		fmt.Printf("🔴 SECURITY UPDATE: %s %s has vulnerability %s\n",
			emblemName, currentVersion, cached.SecurityAdvisory)
		fmt.Printf("   Update to %s immediately! Run: ely update %s\n", cached.LatestVersion, emblemName)
		return
	}

	if cached.LatestVersion != "" && cached.LatestVersion != currentVersion {
		fmt.Printf("⚠️  Outdated: %s@%s (latest: %s)\n", emblemName, currentVersion, cached.LatestVersion)
		fmt.Printf("   Use 'ely update %s' to update\n", emblemName)
	}
}

func init() {
	checkUpdatesCmd.Flags().Bool("no-check", false, "disable update check")
	rootCmd.AddCommand(checkUpdatesCmd)
}
