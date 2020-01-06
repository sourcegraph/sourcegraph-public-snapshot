// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"net/url"
	"reflect"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"gopkg.in/inconshreveable/log15.v2"
)

var externalURLWatchers uint32

var externalURL = func() atomic.Value {
	var v atomic.Value
	v.Store(&url.URL{Scheme: "http", Host: "example.com"})
	return v
}()

// WatchExternalURL watches for changes in the ExternalURL critical config
// so that changes are reflected in what is returned by the ExternalURL function.
// In case the setting is not set, defaultURL is used.
// This should only be called once and will panic otherwise.
func WatchExternalURL(defaultURL *url.URL) {
	if atomic.AddUint32(&externalURLWatchers, 1) != 1 {
		panic("WatchExternalURL called more than once")
	}

	conf.Watch(func() {
		after := defaultURL
		if val := conf.Get().ExternalURL; val != "" {
			var err error
			if after, err = url.Parse(val); err != nil {
				log15.Error("globals.ExternalURL", "value", val, "error", err)
				return
			}
		}

		if before := ExternalURL(); !reflect.DeepEqual(before, after) {
			SetExternalURL(after)
			if before.Host != "example.com" {
				log15.Info(
					"globals.ExternalURL",
					"updated", true,
					"before", before,
					"after", after,
				)
			}
		}
	})
}

// ExternalURL returns the fully-resolved, externally accessible frontend URL.
// Callers must not mutate the returned pointer.
func ExternalURL() *url.URL {
	return externalURL.Load().(*url.URL)
}

// SetExternalURL sets the fully-resolved, externally accessible frontend URL.
func SetExternalURL(u *url.URL) {
	externalURL.Store(u)
}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly *conf.Server
