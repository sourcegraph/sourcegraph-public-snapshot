package httpheader

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// Watch for configuration changes related to the http-header auth provider.
func init() {
	conf.Watch(func() {
		newPC, _ := getProviderConfig()
		if newPC == nil {
			auth.UpdateProviders("httpheader", nil)
		} else {
			auth.UpdateProviders("httpheader", []auth.Provider{&provider{c: newPC}})
		}
	})
}
