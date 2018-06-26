package backend

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

func init() {
	// Use a caching HTTP client for communicating with the remote registry.
	registry.HTTPClient = httputil.CachingClient
}

// GetExtensionByExtensionID gets the extension with the given extension ID.
//
// It returns either a local or remote extension, depending on what the extension ID refers to.
//
// BACKCOMPAT: It also synthesizes registry extensions from known language servers.
func GetExtensionByExtensionID(ctx context.Context, extensionID string) (local *db.RegistryExtension, remote *registry.Extension, err error) {
	if parts := strings.SplitN(extensionID, "/", 3); len(parts) == 3 {
		prefix, err := GetLocalRegistryExtensionIDPrefix()
		if err != nil {
			return nil, nil, err
		}
		if prefix == nil {
			// Don't look up remote extensions from Sourcegraph.com; it only cares about its own
			// extensions.
			return nil, nil, errors.New("remote extension lookup disabled")
		}
		if parts[0] != *prefix {
			return nil, nil, fmt.Errorf("external extension lookup on non-default registry is forbidden (extension ID prefix %q, allowed prefixes are \"\" (default) and %q (local))", parts[0], *prefix)
		}
		x, err := db.RegistryExtensions.GetByExtensionID(ctx, path.Join(parts[1], parts[2]))
		if err != nil {
			return nil, nil, err
		}
		if err := PrefixLocalExtensionID(x); err != nil {
			return nil, nil, err
		}
		return x, nil, nil
	}

	// BACKCOMPAT: Synthesize extensions for known language servers.
	x, err := getSynthesizedRegistryExtension(ctx, "extensionID", extensionID)
	if x != nil || err != nil {
		return nil, x, err
	}

	x, err = GetRemoteRegistryExtension(ctx, "extensionID", extensionID)
	if err != nil {
		return nil, nil, err
	}
	return nil, x, nil
}

// PrefixLocalExtensionID adds the local registry's extension ID prefix (from
// GetLocalRegistryExtensionIDPrefix) to all extensions' extension IDs in the list.
func PrefixLocalExtensionID(xs ...*db.RegistryExtension) error {
	prefix, err := GetLocalRegistryExtensionIDPrefix()
	if err != nil {
		return err
	}
	if prefix == nil {
		return nil
	}
	for _, x := range xs {
		x.NonCanonicalExtensionID = *prefix + "/" + x.NonCanonicalExtensionID
		x.NonCanonicalRegistry = *prefix
	}
	return nil
}

// GetLocalRegistryName returns the name of the local registry.
func GetLocalRegistryName() (string, error) {
	u, err := url.Parse(conf.Get().AppURL)
	if err != nil {
		return "", err
	}
	return registry.Name(u), nil
}

// GetLocalRegistryExtensionIDPrefix returns the extension ID prefix (if any) of extensions in the
// local registry.
func GetLocalRegistryExtensionIDPrefix() (*string, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, nil
	}
	name, err := GetLocalRegistryName()
	if err != nil {
		return nil, nil
	}
	return &name, nil
}

// GetRemoteRegistryURL returns the remote registry URL from site configuration, or nil if there is
// none. If an error exists while parsing the value in site configuration, the error is returned.
func GetRemoteRegistryURL() (*url.URL, error) {
	pc := conf.Platform()
	if pc == nil || pc.RemoteRegistryURL == "" {
		return nil, nil
	}
	return url.Parse(pc.RemoteRegistryURL)
}

// GetRemoteRegistryExtension gets the remote registry extension and rewrites its fields to be from
// the frame-of-reference of this site. The field is either "uuid" or "extensionID".
func GetRemoteRegistryExtension(ctx context.Context, field, value string) (*registry.Extension, error) {
	// BACKCOMPAT: First, look up among extensions synthesized from known language servers.
	if x, err := getSynthesizedRegistryExtension(ctx, field, value); x != nil || err != nil {
		return x, err
	}

	registryURL, err := GetRemoteRegistryURL()
	if registryURL == nil || err != nil {
		return nil, err
	}

	var x *registry.Extension
	switch field {
	case "uuid":
		x, err = registry.GetByUUID(ctx, registryURL, value)
	case "extensionID":
		x, err = registry.GetByExtensionID(ctx, registryURL, value)
	default:
		panic("unexpected field: " + field)
	}
	if x != nil {
		x.RegistryURL = registryURL.String()
	}
	return x, err
}

// ListRemoteRegistryExtensions lists the remote registry extensions and rewrites their fields to be
// from the frame-of-reference of this site.
func ListRemoteRegistryExtensions(ctx context.Context, query string) ([]*registry.Extension, error) {
	registryURL, err := GetRemoteRegistryURL()
	if registryURL == nil || err != nil {
		return nil, err
	}

	xs, err := registry.List(ctx, registryURL, query)
	if err != nil {
		return nil, err
	}
	return xs, nil
}
