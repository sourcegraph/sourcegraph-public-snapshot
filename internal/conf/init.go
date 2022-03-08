package conf

import (
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// Init function must be called by every service that requires access to components in conf package.
// If the caller is in the client mode, we start a goroutine continuously polling for config changes.
func Init() {
	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if getMode() == modeClient {
		go DefaultClient().continuouslyUpdate(nil)
		close(configurationServerFrontendOnlyInitialized)
	}

	// This watch loop is here so that we don't introduce
	// package dependency cycles, since conf itself uses httpcli's internal
	// client. This is gross, and the whole conf package is gross.
	go Watch(func() {
		before := httpcli.TLSExternalConfig()
		after := Get().ExperimentalFeatures.TlsExternal
		if !reflect.DeepEqual(before, after) {
			httpcli.SetTLSExternalConfig(after)
		}
	})
}
