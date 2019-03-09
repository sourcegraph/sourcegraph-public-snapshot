package userpasswd

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// Watch for configuration changes related to the builtin auth provider.
func init() {
	go func() {
		conf.Watch(func() {
			newPC, _ := getProviderConfig()
			if newPC == nil {
				providers.UpdateProviders("builtin", nil)
				return
			}
			providers.UpdateProviders("builtin", []providers.Provider{&provider{c: newPC}})
		})
	}()
}
