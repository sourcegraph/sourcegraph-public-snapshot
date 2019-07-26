package httpheader

import (
	"sourcegraph.com/cmd/frontend/auth/providers"
	"sourcegraph.com/pkg/conf"
)

// Watch for configuration changes related to the http-header auth provider.
func init() {
	go func() {
		conf.Watch(func() {
			newPC, _ := getProviderConfig()
			if newPC == nil {
				providers.Update("httpheader", nil)
				return
			}
			providers.Update("httpheader", []providers.Provider{&provider{c: newPC}})
		})
	}()
}
