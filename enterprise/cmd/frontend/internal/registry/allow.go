package registry

import (
	"strings"

	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	frontendregistry.IsRemoteExtensionAllowed = func(extensionID string) bool {
		if getAllowOnlySourcegraphAuthoredExtensionsFromSiteConfig() && !strings.HasPrefix(extensionID, "sourcegraph/") {
			return false
		}

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
		var extensionsWithAllowedPublisher []*registry.Extension

		if getAllowOnlySourcegraphAuthoredExtensionsFromSiteConfig() {
			for _, x := range extensions {
				if strings.HasPrefix(x.ExtensionID, "sourcegraph/") {
					extensionsWithAllowedPublisher = append(extensionsWithAllowedPublisher, x)
				}
			}
		} else {
			// If allow only Sourcegraph extensions feature is not enabled, all remote extensions are allowed.
			extensionsWithAllowedPublisher = extensions
		}

		allowedExtensionIDs := getAllowedExtensionsFromSiteConfig()

		if allowedExtensionIDs == nil {
			return extensionsWithAllowedPublisher
		}

		allow := make(map[string]any)
		for _, id := range allowedExtensionIDs {
			allow[id] = struct{}{}
		}
		var keep []*registry.Extension
		for _, x := range extensionsWithAllowedPublisher {
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
	if !licensing.IsFeatureEnabledLenient(licensing.FeatureRemoteExtensionsAllowDisallow) {
		return nil
	}

	if c := conf.Get().Extensions; c != nil {
		return c.AllowRemoteExtensions
	}
	return nil
}

func getAllowOnlySourcegraphAuthoredExtensionsFromSiteConfig() bool {
	if c := conf.Get().Extensions; c != nil {
		return c.AllowOnlySourcegraphAuthoredExtensions
	}

	return false
}
