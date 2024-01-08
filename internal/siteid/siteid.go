// Package siteid provides access to the site ID, a stable identifier for the current
// Sourcegraph site.
//
// All servers that are part of the same logical Sourcegraph site have the same site ID
// (although it is possible for an admin to misconfigure the servers so that this is not
// true).
//
// The "site ID" was formerly known as the "app ID".
package siteid

import (
	"context"
	"log" //nolint:logging // TODO move all logging to sourcegraph/log
	"os"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

var (
	initOnce sync.Once
	siteID   string

	fatalln = log.Fatalln // overridden in tests
)

// get reads (or generates) the site ID
func get(db database.DB) string {
	if v := os.Getenv("TRACKING_APP_ID"); v != "" {
		// Legacy way of specifying site ID.
		//
		// TODO(dadlerj): remove this
		return v
	} else {
		// Site ID is retrieved from the database (where it might be created automatically
		// if it doesn't yet exist.)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		globalState, err := db.GlobalState().Get(ctx)
		if err != nil {
			fatalln("Error initializing global state:", err)
		}
		return globalState.SiteID
	}
}

// Get returns the site ID.
func Get(db database.DB) string {
	initOnce.Do(func() {
		siteID = get(db)
	})
	return siteID
}
