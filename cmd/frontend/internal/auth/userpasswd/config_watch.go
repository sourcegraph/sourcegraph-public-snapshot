package userpasswd

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// Watch for configuration changes related to the builtin auth provider.
func init() {
	go func() {
		conf.Watch(func() {
			newPC, _ := getProviderConfig()
			if newPC == nil {
				auth.UpdateProviders("builtin", nil)
				return
			}
			auth.UpdateProviders("builtin", []auth.Provider{&provider{c: newPC}})
		})
	}()
}
