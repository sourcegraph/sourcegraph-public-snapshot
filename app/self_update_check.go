package app

import (
	"log"
	"sync"
	"time"

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
			log.Printf("During checking for updates: %s\n", err)
		}

		// Grab the mutex and set the available update version string (only if the
		// current version is not the latest version).
		availableUpdate.Lock()
		if u.CurrentVersion != u.Info.Version {
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
