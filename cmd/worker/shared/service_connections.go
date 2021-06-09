package shared

import (
	"log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// WatchServiceConnectionValue returns the value returned by the given function when passed the
// current service connection configuration. If this function returns a different value in the
// future for an updated service connection configuration, a fatal log will be emitted to
// restart the service to pick up changes.
//
// This method should only be called for critical values like database connection config.
func WatchServiceConnectionValue(f func(serviceConnections conftypes.ServiceConnections) string) string {
	value := f(conf.Get().ServiceConnections)
	conf.Watch(func() {
		if newValue := f(conf.Get().ServiceConnections); value != newValue {
			log.Fatalf("Detected settings change change, restarting to take effect: %s", newValue)
		}
	})

	return value
}
