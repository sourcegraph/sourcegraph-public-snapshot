package registry

import (
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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

	frontendregistry.IsRemoteExtensionPublisherAllowed = func(p registry.Publisher) bool {
		if getAllowOnlySourcegraphAuthoredExtensionsFromSiteConfig() {
			return isSourcegraphAuthoredExtension(p)
		}

		return true
	}

	frontendregistry.FilterRemoteExtensions = func(extensions []*registry.Extension) []*registry.Extension {
		var keep []*registry.Extension

		allowedExtensions := getAllowedExtensionsFromSiteConfig()
		if allowedExtensions != nil {
			allow := make(map[string]struct{})
			for _, id := range allowedExtensions {
				allow[id] = struct{}{}
			}
			for _, x := range extensions {
				if _, ok := allow[x.ExtensionID]; ok {
					keep = append(keep, x)
				}
			}
			return keep
		}

		if getAllowOnlySourcegraphAuthoredExtensionsFromSiteConfig() {
			for _, x := range extensions {
				if isSourcegraphAuthoredExtension(x.Publisher) {
					keep = append(keep, x)
				}
			}
			return keep
		}

		return extensions
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
	if c := conf.Get().Extensions; c != nil && (c.RemoteRegistry == nil || c.RemoteRegistry == conf.DefaultRemoteRegistry) {
		return c.AllowOnlySourcegraphAuthoredExtensions
	}

	return false
}

func isSourcegraphAuthoredExtension(p registry.Publisher) bool {
	return p.Name == "sourcegraph"
}
