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
	for {
		// Check for updates.
		if err := u.Check(); err != nil {
			log15.Warn("error checking for updates", "error", err)
			continue
		}

		// Parse version strings.
		currentVersion, err := version.NewVersion(u.CurrentVersion)
		if err != nil {
			log15.Warn("error parsing current version", "error", err)
			continue
		}
		latestVersion, err := version.NewVersion(u.Info.Version)
		if err != nil {
			log15.Warn("error parsing latest version", "error", err)
			continue
		}

		// If our version string has prelease information (`-suffix`), then this is
		// a private version. Do not check for updates.
		if len(currentVersion.Prerelease()) > 0 {
			log15.Info("disable update check (found private version)", "version", currentVersion)
			return
		}

		// Grab the mutex and set the available update version string (only if the
		// current version is less than the latest version).
		availableUpdate.Lock()
		if currentVersion.LessThan(latestVersion) {
			availableUpdate.Version = u.Info.Version
		}
		availableUpdate.Unlock()

		// Wait a good duration before checking again.
		time.Sleep(appconf.Flags.CheckForUpdates)
	}
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
