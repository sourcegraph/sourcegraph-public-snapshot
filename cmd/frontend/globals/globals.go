// Package globals contains global variables that should be set by the frontend's main function on initialization.
package globals

import (
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var defaultExternalURL = &url.URL{
	Scheme: "http",
	Host:   "example.com",
}

var externalURL = func() atomic.Value {
	var v atomic.Value
	v.Store(defaultExternalURL)
	return v
}()

var watchExternalURLOnce sync.Once

// WatchExternalURL watches for changes in the `externalURL` site configuration
// so that changes are reflected in what is returned by the ExternalURL function.
func WatchExternalURL() {
	logger := log.Scoped("externalURLWatcher", "")

	watchExternalURLOnce.Do(func() {
		conf.Watch(func() {
			after := defaultExternalURL
			if val := conf.Get().ExternalURL; val != "" {
				var err error
				if after, err = url.Parse(val); err != nil {
					logger.Error("globals.ExternalURL", log.String("value", val), log.Error(err))
					return
				}
			}

			if before := ExternalURL(); !reflect.DeepEqual(before, after) {
				SetExternalURL(after)
				if before.Host != "example.com" {
					logger.Info(
						"globals.ExternalURL",
						log.Bool("updated", true),
						log.String("before", before.String()),
						log.String("after", after.String()),
					)
				}
			}
		})
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

var defaultPermissionsUserMapping = &schema.PermissionsUserMapping{
	Enabled: false,
	BindID:  "email",
}

// permissionsUserMapping mirrors the value of `permissions.userMapping` in the site configuration.
// This variable is used to monitor configuration change via conf.Watch and must be operated atomically.
var permissionsUserMapping = func() atomic.Value {
	var v atomic.Value
	v.Store(defaultPermissionsUserMapping)
	return v
}()

var watchPermissionsUserMappingOnce sync.Once

// WatchPermissionsUserMapping watches for changes in the `permissions.userMapping` site configuration
// so that changes are reflected in what is returned by the PermissionsUserMapping function.
func WatchPermissionsUserMapping() {
	logger := log.Scoped("permissionsUserMapWatcher", "")

	watchPermissionsUserMappingOnce.Do(func() {
		conf.Watch(func() {
			after := conf.Get().PermissionsUserMapping
			if after == nil {
				after = defaultPermissionsUserMapping
			} else if after.BindID != "email" && after.BindID != "username" {
				logger.Error("globals.PermissionsUserMapping", log.String("BindID", after.BindID), log.String("error", "not a valid value"))
				return
			}

			if before := PermissionsUserMapping(); !reflect.DeepEqual(before, after) {
				SetPermissionsUserMapping(after)
				logger.Info(
					"globals.PermissionsUserMapping",
					log.Bool("updated", true),
					log.String("before", fmt.Sprint(before)),
					log.String("after", fmt.Sprint(after)),
				)
			}
		})
	})
}

// PermissionsUserMapping returns the last valid value of permissions user mapping in the site configuration.
// Callers must not mutate the returned pointer.
func PermissionsUserMapping() *schema.PermissionsUserMapping {
	return permissionsUserMapping.Load().(*schema.PermissionsUserMapping)
}

// SetPermissionsUserMapping sets a valid value for the permissions user mapping.
func SetPermissionsUserMapping(u *schema.PermissionsUserMapping) {
	permissionsUserMapping.Store(u)
}

var defaultBranding = &schema.Branding{
	BrandName: "Sourcegraph",
}

// branding mirrors the value of `branding` in the site configuration.
// This variable is used to monitor configuration change via conf.Watch and must be operated atomically.
var branding = func() atomic.Value {
	var v atomic.Value
	v.Store(defaultBranding)
	return v
}()

var brandingWatchers uint32

// WatchBranding watches for changes in the `branding` site configuration
// so that changes are reflected in what is returned by the Branding function.
// This should only be called once and will panic otherwise.
func WatchBranding() {
	if atomic.AddUint32(&brandingWatchers, 1) != 1 {
		panic("WatchBranding called more than once")
	}

	logger := log.Scoped("brandingWatcher", "")

	conf.Watch(func() {
		after := conf.Get().Branding
		if after == nil {
			after = defaultBranding
		} else if after.BrandName == "" {
			bcopy := *after
			bcopy.BrandName = defaultBranding.BrandName
			after = &bcopy
		}

		if before := Branding(); !reflect.DeepEqual(before, after) {
			SetBranding(after)
			logger.Debug(
				"globals.Branding",
				log.Bool("updated", true),
				log.String("before", fmt.Sprint(before)),
				log.String("after", fmt.Sprint(after)),
			)
		}
	})
}

// Branding returns the last valid value of branding in the site configuration.
// Callers must not mutate the returned pointer.
func Branding() *schema.Branding {
	return branding.Load().(*schema.Branding)
}

// SetBranding sets a valid value for the branding.
func SetBranding(u *schema.Branding) {
	branding.Store(u)
}

// ConfigurationServerFrontendOnly provides the contents of the site configuration
// to other services and manages modifications to it.
//
// Any another service that attempts to use this variable will panic.
var ConfigurationServerFrontendOnly *conf.Server
