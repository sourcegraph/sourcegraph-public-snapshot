package registry

import (
	"github.com/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

func init() {
	frontendregistry.IsRemoteExtensionAllowed = func(extensionID string) bool {
		allowedExtensions := getAllowedExtensionsFromSiteConfig()
		if allowedExtensions == nil {
			// Default is to allow all extensions.
			return true
		}

		for _, x := range allowedExtensions {
			if extensionID == x {
				return true
			}
		}
		return false
	}

	frontendregistry.FilterRemoteExtensions = func(extensions []*registry.Extension) []*registry.Extension {
		allowedExtensions := getAllowedExtensionsFromSiteConfig()
		if allowedExtensions == nil {
			// Default is to allow all extensions.
			return extensions
		}

		allow := make(map[string]interface{})
		for _, id := range allowedExtensions {
			allow[id] = struct{}{}
		}
		var keep []*registry.Extension
		for _, x := range extensions {
			if _, ok := allow[x.ExtensionID]; ok {
				keep = append(keep, x)
			}
		}
		return keep
	}
}

func getAllowedExtensionsFromSiteConfig() []string {
	// If the remote extension allow/disallow feature is not enabled, all remote extensions are
	// allowed. This is achieved by a nil list.
	if !isLicenseFeatureEnabledLenient(licensing.FeatureRemoteExtensionsAllowDisallow) {
		return nil
	}

	if c := conf.Get().Extensions; c != nil {
		return c.AllowRemoteExtensions
	}
	return nil
}

// isLicenseFeatureEnabledLenient reports whether the current license enables the given feature. If
// there is an error reading the license, it is lenient and returns true.
//
// This is useful for callers who don't want to handle errors (usually because the user would be
// prevented from getting to this point if license verification had failed, so it's not necessary to
// handle license verification errors here).
func isLicenseFeatureEnabledLenient(feature licensing.Feature) bool {
	return !licensing.IsFeatureNotActivated(licensing.CheckFeature(feature))
}
