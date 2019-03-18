package registry

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"

	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

func init() {
	// Use a caching HTTP client for communicating with the remote registry.
	const sleepIfUncached = false
	if env.InsecureDev && sleepIfUncached {
		// Also simulate latency in dev mode. See docs for sleepIfUncachedTransport for more information.
		registry.HTTPClient = &http.Client{Transport: sleepIfUncachedTransport{httputil.CachingClient.Transport}}
	} else {
		registry.HTTPClient = httputil.CachingClient
	}
}

// SplitExtensionID splits an extension ID of the form [host/]publisher/name (where [host/] is the
// optional registry prefix), such as "alice/myextension" or
// "sourcegraph.example.com/bob/myextension". It returns the components, or a non-nil error if
// parsing failed.
func SplitExtensionID(extensionID string) (prefix, publisher, name string, err error) {
	parts := strings.Split(extensionID, "/")
	if len(parts) == 0 || len(parts) == 1 {
		return "", "", "", fmt.Errorf("invalid extension ID: %q (2+ slash-separated path components required)", extensionID)
	}
	name = parts[len(parts)-1] // last
	if name == "" {
		return "", "", "", fmt.Errorf("invalid extension ID: %q (trailing slash is forbidden)", extensionID)
	}
	publisher = parts[len(parts)-2] // 2nd to last
	if publisher == "" {
		return "", "", "", fmt.Errorf("invalid extension ID: %q (empty publisher)", extensionID)
	}
	prefix = strings.Join(parts[:len(parts)-2], "/") // prefix
	return
}

// ParseExtensionID parses an extension ID of the form [host/]publisher/name (where [host/] is the
// optional registry prefix), such as "alice/myextension" or
// "sourcegraph.example.com/bob/myextension". It validates that the registry prefix is correct given
// the current configuration.
func ParseExtensionID(extensionID string) (prefix, extensionIDWithoutPrefix string, isLocal bool, err error) {
	prefix, publisher, name, err := SplitExtensionID(extensionID)
	if err != nil {
		return "", "", false, err
	}

	configuredPrefix := GetLocalRegistryExtensionIDPrefix()
	if prefix != "" {
		// Extension ID is host/publisher/name.
		if configuredPrefix == nil {
			// Don't look up fully qualified extensions from Sourcegraph.com; it only cares about
			// its own extensions.
			return "", "", false, fmt.Errorf("remote extension lookup is not supported for host %q", prefix)
		}

		// Local extension on non-Sourcegraph.com instance.
		if prefix != *configuredPrefix {
			return "", "", false, fmt.Errorf("remote extension lookup is forbidden (extension ID prefix %q, allowed prefixes are \"\" (default) and %q (local))", prefix, *configuredPrefix)
		}
		isLocal = true
	} else {
		// Extension ID is publisher/name.
		if configuredPrefix == nil {
			// Local extension on Sourcegraph.com instance.
			isLocal = true
		}
	}

	extensionIDWithoutPrefix = publisher + "/" + name
	return prefix, extensionIDWithoutPrefix, isLocal, nil
}

// GetLocalExtensionByExtensionID looks up and returns the registry extension in the local registry
// with the given extension ID. If there is no local extension registry, it is not implemented.
var GetLocalExtensionByExtensionID func(ctx context.Context, extensionIDWithoutPrefix string) (local graphqlbackend.RegistryExtension, err error)

// GetExtensionByExtensionID gets the extension with the given extension ID.
//
// It returns either a local or remote extension, depending on what the extension ID refers to.
//
// The format of an extension ID is [host/]publisher/name. If the host is omitted, the host defaults
// to the remote registry specified in site configuration (usually sourcegraph.com). The host must
// be specified to refer to a local extension on the current Sourcegraph site (e.g.,
// sourcegraph.example.com/publisher/name).
//
// BACKCOMPAT: It also synthesizes registry extensions from known language servers.
func GetExtensionByExtensionID(ctx context.Context, extensionID string) (local graphqlbackend.RegistryExtension, remote *registry.Extension, err error) {
	_, extensionIDWithoutPrefix, isLocal, err := ParseExtensionID(extensionID)
	if err != nil {
		return nil, nil, err
	}

	if isLocal {
		// BACKCOMPAT: First, look up among extensions synthesized from known language servers.
		if x, err := getSynthesizedRegistryExtension(ctx, "extensionID", extensionID); x != nil || err != nil {
			return nil, x, err
		}

		if GetLocalExtensionByExtensionID != nil {
			x, err := GetLocalExtensionByExtensionID(ctx, extensionIDWithoutPrefix)
			return x, nil, err
		}
	}

	x, err := getRemoteRegistryExtension(ctx, "extensionID", extensionIDWithoutPrefix)
	if err != nil {
		return nil, nil, err
	}
	return nil, x, nil
}

// getLocalRegistryName returns the name of the local registry.
func getLocalRegistryName() string {
	u, err := url.Parse(conf.Get().AppURL)
	if err != nil || u == nil || u.Host == "" {
		return registry.Name(globals.AppURL)
	}
	return registry.Name(u)
}

var mockLocalRegistryExtensionIDPrefix **string

// GetLocalRegistryExtensionIDPrefix returns the extension ID prefix (if any) of extensions in the
// local registry.
func GetLocalRegistryExtensionIDPrefix() *string {
	if mockLocalRegistryExtensionIDPrefix != nil {
		return *mockLocalRegistryExtensionIDPrefix
	}
	if envvar.SourcegraphDotComMode() {
		return nil
	}
	name := getLocalRegistryName()
	return &name
}

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

var mockGetRemoteRegistryExtension func(field, value string) (*registry.Extension, error)

// getRemoteRegistryExtension gets the remote registry extension and rewrites its fields to be from
// the frame-of-reference of this site. The field is either "uuid" or "extensionID".
func getRemoteRegistryExtension(ctx context.Context, field, value string) (*registry.Extension, error) {
	if mockGetRemoteRegistryExtension != nil {
		return mockGetRemoteRegistryExtension(field, value)
	}

	// BACKCOMPAT: First, look up among extensions synthesized from known language servers.
	if x, err := getSynthesizedRegistryExtension(ctx, field, value); x != nil || err != nil {
		return x, err
	}

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

	if !IsRemoteExtensionAllowed(x.ExtensionID) {
		return nil, fmt.Errorf("extension is not allowed in site configuration: %q", x.ExtensionID)
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

// sleepIfUncachedTransport is used to simulate latency in local dev mode.
//
// This helps us feel the UX of not being in Northern California latency-wise and ensure that
// Sourcegraph's communication with the remote extension registry (usually Sourcegraph.com) does not
// block unrelated workflows.
type sleepIfUncachedTransport struct {
	http.RoundTripper
}

func (t sleepIfUncachedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil || resp.Header.Get(httpcache.XFromCache) == "" {
		n := rand.Intn(750)
		time.Sleep(time.Duration(750+n) * time.Millisecond)
	}
	return resp, err
}
