package conf

import (
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func Init() {
	// The default client is started in InitConfigurationServerFrontendOnly in
	// the case of server mode.
	if getMode() == modeClient {
		go defaultClientVal.continuouslyUpdate(nil)
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
