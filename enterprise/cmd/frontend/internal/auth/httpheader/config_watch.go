package httpheader

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// Watch for configuration changes related to the http-header auth provider.
func init() {
	go func() {
		const pkgName = "httpheader"
		logger := log.Scoped(pkgName, "HTTP header authentication config watch")
		conf.Watch(func() {
			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Warn("Check license for SSO (HTTP header)", log.Error(err))
				providers.Update(pkgName, nil)
				return
			}

			newPC, _ := getProviderConfig()
			if newPC == nil {
				providers.Update(pkgName, nil)
				return
			}
			providers.Update(pkgName, []providers.Provider{&provider{c: newPC}})
		})
	}()
}
