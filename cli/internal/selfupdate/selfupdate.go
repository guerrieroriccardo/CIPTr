package selfupdate

import (
	"context"
	"fmt"
	"os"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/version"
)

func Run() {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update source error: %v\n", err)
		os.Exit(1)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Updater error: %v\n", err)
		os.Exit(1)
	}

	latest, found, err := updater.DetectLatest(
		context.Background(),
		selfupdate.NewRepositorySlug("guerrieroriccardo", "CIPTr"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
		os.Exit(1)
	}
	if !found {
		fmt.Println("No updates available.")
		return
	}
	if !latest.GreaterThan(version.Version) {
		fmt.Printf("Already up to date (%s).\n", version.Version)
		return
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot find executable path: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updating %s → %s...\n", version.Version, latest.Version())
	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Updated successfully!")
}
