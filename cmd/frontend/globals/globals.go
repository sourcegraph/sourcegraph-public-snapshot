// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"net/url"
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	externalURLWatch sync.Once
	externalURLMutex sync.RWMutex
	externalURL      = &url.URL{Scheme: "http", Host: "example.com"}
)

// WatchExternalURL watches for changes in the ExternalURL critical config
// so that changes are reflected in what is returned by the ExternalURL function.
// In case the setting is not set, defaultURL is used.
func WatchExternalURL(defaultURL *url.URL) {
	externalURLWatch.Do(func() {
		conf.Watch(func() {
			after := defaultURL
			if val := conf.Get().Critical.ExternalURL; val != "" {
				var err error
				if after, err = url.Parse(val); err != nil {
					log15.Error("globals.ExternalURL", "value", val, "error", err)
					return
				}
			}

			externalURLMutex.Lock()
			defer externalURLMutex.Unlock()

			before := *externalURL
			if !reflect.DeepEqual(&before, after) {
				externalURL = after
				log15.Info(
					"globals.ExternalURL",
					"updated", true,
					"before", &before,
					"after", after,
				)
			}
		})
	})
}

// ExternalURL returns the fully-resolved, externally accessible frontend URL.
func ExternalURL() *url.URL {
	externalURLMutex.RLock()
	defer externalURLMutex.RUnlock()

	u := *externalURL
	if u.User != nil {
		user := *(u.User)
		u.User = &user
	}

	return &u
}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly *conf.Server
