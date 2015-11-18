package app

import (
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/hashicorp/go-version"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

// Stores the latest available update version string (or an empty string if
// there is none).
var availableUpdate struct {
	sync.RWMutex
	Version string
}

// checkForUpdates runs in a separate goroutine and periodically checks for
// updates to the application.
func checkForUpdates() {
	u := sgxcmd.SelfUpdater

	// Parse our current version string.
	currentVersion, err := version.NewVersion(u.CurrentVersion)
	if err != nil {
		if u.CurrentVersion == "dev" {
			return
		}
		log15.Info("disable update check (found invalid version)", "version", u.CurrentVersion)
		return
	}

	// If our version string has prelease information (`-suffix`), then this is
	// a private version. Do not check for updates.
	if len(currentVersion.Prerelease()) > 0 {
		log15.Info("disable update check (found private version)", "version", currentVersion)
		return
	}

	for {
		// Check for updates once.
		checkForUpdateOnce(currentVersion)

		// Wait a good duration before checking again.
		time.Sleep(appconf.Flags.CheckForUpdates)
	}
}

// checkForUpdateOnce checks for an update once.
func checkForUpdateOnce(currentVersion *version.Version) {
	u := sgxcmd.SelfUpdater

	// Check for updates.
	if err := u.Check(); err != nil {
		log15.Warn("error checking for updates", "error", err)
		return
	}

	// Parse the latest binary version string.
	latestVersion, err := version.NewVersion(u.Info.Version)
	if err != nil {
		log15.Warn("error parsing latest version", "error", err)
		return
	}

	// Grab the mutex and set the available update version string (only if the
	// current version is less than the latest version).
	availableUpdate.Lock()
	if currentVersion.LessThan(latestVersion) {
		availableUpdate.Version = u.Info.Version
	}
	availableUpdate.Unlock()
}

func init() {
	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		if appconf.Flags.CheckForUpdates != 0 {
			go checkForUpdates()
		}
	})
}

// updateAvailable returns the version string of an available updated, or
// returns an empty string if no update is available. It is safe to call from
// multiple goroutines concurrently.
func updateAvailable() string {
	availableUpdate.Lock()
	v := availableUpdate.Version
	availableUpdate.Unlock()
	return v
}
