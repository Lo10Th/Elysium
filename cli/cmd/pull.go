package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <emblem-name>[@version] [emblem-name[@version]...]",
	Short: "Download and cache one or more emblems",
	Long: `Download one or more emblems from the registry and cache them locally.

Examples:
  # Pull the latest version
  ely pull clothing-shop
  
  # Pull a specific version
  ely pull clothing-shop@1.0.0
  
  # Pull multiple emblems at once
  ely pull clothing-shop stripe github slack`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		if len(args) == 1 {
			return pullSingleEmblem(args[0], verbose)
		}

		fmt.Printf("✓ Pulling %d emblems...\n\n", len(args))

		type result struct {
			arg string
			err error
		}

		results := make(chan result, len(args))
		var wg sync.WaitGroup

		for _, arg := range args {
			wg.Add(1)
			go func(a string) {
				defer wg.Done()
				err := pullSingleEmblem(a, verbose)
				results <- result{arg: a, err: err}
			}(arg)
		}

		wg.Wait()
		close(results)

		var failures []string
		for r := range results {
			name := strings.Split(r.arg, "@")[0]
			if r.err != nil {
				fmt.Printf("%s:\n  ✗ %v\n", name, r.err)
				failures = append(failures, name)
			}
		}

		if len(failures) > 0 {
			return fmt.Errorf("%d emblem(s) failed to pull: %s", len(failures), strings.Join(failures, ", "))
		}

		return nil
	},
}

// pullSingleEmblem downloads a single emblem by name (with optional @version).
func pullSingleEmblem(arg string, verbose bool) error {
	name, version, err := emblem.ParseVersionConstraint(arg)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("Pulling emblem %s@%s...\n", name, version)
	}

	cfg := config.Get()
	client := api.NewClient()
	client.SetToken(cfg.Token)

	emblemInfo, err := client.GetEmblem(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errfmt.EmblemNotFoundError(name)
		}
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return errfmt.ConnectionError(cfg.Registry, err)
		}
		if strings.Contains(err.Error(), "timeout") {
			return errfmt.NewDetailedError(err).
				WithReason("Request timed out").
				WithSuggestion("Try again or check your network connection")
		}
		return errfmt.NetworkError(err)
	}

	if version == "latest" {
		version = emblemInfo.LatestVersion
	}

	ver, err := client.GetEmblemVersion(name, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errfmt.NewDetailedError(fmt.Errorf("version '%s' not found for emblem '%s'", version, name)).
				WithSuggestion(fmt.Sprintf("Try: ely pull %s (to get latest version)", name))
		}
		if strings.Contains(err.Error(), "connection refused") {
			return errfmt.ConnectionError(cfg.Registry, err)
		}
		return errfmt.NetworkError(err)
	}

	if err := emblem.SaveToCache(name, version, []byte(ver.YAMLContent)); err != nil {
		return fmt.Errorf("failed to cache emblem: %w", err)
	}

	if err := config.InstallEmblem(name, version); err != nil {
		return fmt.Errorf("failed to update installed list: %w", err)
	}

	// Cache the latest version info so update notifications work offline.
	if cacheErr := config.SetVersionCache(
		name,
		emblemInfo.LatestVersion,
		emblemInfo.SecurityAdvisory,
		emblemInfo.SecuritySeverity,
	); cacheErr != nil {
		// Non-fatal; don't fail the pull.
		fmt.Printf("Warning: could not cache version info for %s: %v\n", name, cacheErr)
	}

	fmt.Printf("%s:\n  ✓ Downloaded %s@%s\n", name, name, version)

	// Notify if the pulled version is behind the latest.
	if version != emblemInfo.LatestVersion {
		fmt.Printf("  ⚠️  Update available: %s %s → %s\n", name, version, emblemInfo.LatestVersion)
		fmt.Printf("     Run: ely update %s\n", name)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
