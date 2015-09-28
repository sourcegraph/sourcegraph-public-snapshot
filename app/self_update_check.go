package app

import (
	"log"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/sgxcmd"
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
			log.Printf("During checking for updates: %s\n", err)
		}

		// Grab the mutex and update the latest version string.
		availableUpdate.Lock()
		availableUpdate.Version = u.Info.Version
		availableUpdate.Unlock()

		// Wait a good duration before checking again.
		time.Sleep(appconf.Current.CheckForUpdates)
	}
}

func init() {
	if appconf.Current.CheckForUpdates != 0 {
		go checkForUpdates()
	}
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
