package userpasswd

import (
	"sourcegraph.com/cmd/frontend/auth/providers"
	"sourcegraph.com/pkg/conf"
)

// Watch for configuration changes related to the builtin auth provider.
func init() {
	go func() {
		conf.Watch(func() {
			newPC, _ := getProviderConfig()
			if newPC == nil {
				providers.Update("builtin", nil)
				return
			}
			providers.Update("builtin", []providers.Provider{&provider{c: newPC}})
		})
	}()
}
