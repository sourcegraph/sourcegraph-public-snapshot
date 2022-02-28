package conf

import (
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func Init() {
	// This watch loop is here so that we don't introduce
	// dependency cycles, since conf itself uses httpcli's internal
	// client. This is gross, and the whole conf package is gross.
	go Watch(func() {
		before := httpcli.TLSExternalConfig()
		after := Get().ExperimentalFeatures.TlsExternal
		if !reflect.DeepEqual(before, after) {
			httpcli.SetTLSExternalConfig(after)
		}
	})

	Ready()
}
