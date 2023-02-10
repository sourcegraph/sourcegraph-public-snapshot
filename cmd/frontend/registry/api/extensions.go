package api

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// getRemoteRegistryURL returns the remote registry URL from site configuration, or nil if there is
// none. If an error exists while parsing the value in site configuration, the error is returned.
func getRemoteRegistryURL() (*url.URL, error) {
	pc := conf.Extensions()
	if pc == nil || pc.RemoteRegistryURL == "" {
		return nil, nil
	}
	return url.Parse(pc.RemoteRegistryURL)
}

// IsRemoteExtensionAllowed reports whether to allow usage of the remote extension with the given
// extension ID.
//
// It can be overridden to use custom logic.
var IsRemoteExtensionAllowed = func(extensionID string) bool {
	// By default, all remote extensions are allowed.
	return true
}

// IsRemoteExtensionPublisherAllowed reports whether to allow usage of the remote extension created by
// certain publisher by extension ID.
//
// It can be overridden to use custom logic.
var IsRemoteExtensionPublisherAllowed = func(p registry.Publisher) bool {
	// By default, all remote extensions are allowed.
	return true
}

// getRemoteRegistryExtension gets the remote registry extension and rewrites its fields to be from
// the frame-of-reference of this site. The field is either "uuid" or "extensionID".
func getRemoteRegistryExtension(ctx context.Context, field, value string) (*registry.Extension, error) {
	registryURL, err := getRemoteRegistryURL()
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

	if x != nil && !IsRemoteExtensionAllowed(x.ExtensionID) {
		return nil, errors.Errorf("extension is not allowed in site configuration: %q", x.ExtensionID)
	}

	if x != nil && !IsRemoteExtensionPublisherAllowed(x.Publisher) {
		return nil, errors.Errorf("Only extensions authored by Sourcegraph are allowed in this site configuration")
	}

	return x, err
}

// FilterRemoteExtensions is called to filter the list of extensions retrieved from the remote
// registry before the list is used by any other part of the application.
//
// It can be overridden to use custom logic.
var FilterRemoteExtensions = func(extensions []*registry.Extension) []*registry.Extension {
	// By default, all remote extensions are allowed.
	return extensions
}

// listRemoteRegistryExtensions lists the remote registry extensions and rewrites their fields to be
// from the frame-of-reference of this site.
func listRemoteRegistryExtensions(ctx context.Context, query string) ([]*registry.Extension, error) {
	registryURL, err := getRemoteRegistryURL()
	if registryURL == nil || err != nil {
		return nil, err
	}

	xs, err := registry.List(ctx, registryURL, query)
	if err != nil {
		return nil, err
	}
	xs = FilterRemoteExtensions(xs)
	for _, x := range xs {
		x.RegistryURL = registryURL.String()
	}
	return xs, nil
}
