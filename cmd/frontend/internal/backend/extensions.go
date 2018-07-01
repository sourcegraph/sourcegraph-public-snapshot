package backend

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

func init() {
	// Use a caching HTTP client for communicating with the remote registry.
	if envvar.InsecureDevMode() {
		// Also simulate latency in dev mode. See docs for sleepIfUncachedTransport for more information.
		registry.HTTPClient = &http.Client{Transport: sleepIfUncachedTransport{httputil.CachingClient.Transport}}
	} else {
		registry.HTTPClient = httputil.CachingClient
	}
}

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
func GetExtensionByExtensionID(ctx context.Context, extensionID string) (local *db.RegistryExtension, remote *registry.Extension, err error) {
	parts := strings.SplitN(extensionID, "/", 3)
	if len(parts) == 0 || len(parts) == 1 {
		return nil, nil, fmt.Errorf("invalid extension ID: %q", extensionID)
	}

	prefix, err := GetLocalRegistryExtensionIDPrefix()
	if err != nil {
		return nil, nil, err
	}

	var isLocal bool
	switch len(parts) {
	case 3: // extension ID is host/publisher/name
		if prefix == nil {
			// Don't look up fully qualified extensions from Sourcegraph.com; it only cares about
			// its own extensions.
			return nil, nil, errors.New("remote extension lookup disabled")
		}

		// Local extension on non-Sourcegraph.com instance.
		if parts[0] != *prefix {
			return nil, nil, fmt.Errorf("external extension lookup on non-default registry is forbidden (extension ID prefix %q, allowed prefixes are \"\" (default) and %q (local))", parts[0], *prefix)
		}
		extensionID = path.Join(parts[1], parts[2])
		isLocal = true

	case 2: // extension ID is publisher/name
		if prefix == nil {
			// Local extension on Sourcegraph.com instance.
			isLocal = true
		}
	}

	if isLocal {
		x, err := db.RegistryExtensions.GetByExtensionID(ctx, extensionID)
		if err != nil {
			return nil, nil, err
		}
		if err := PrefixLocalExtensionID(x); err != nil {
			return nil, nil, err
		}
		return x, nil, nil
	}

	x, err := GetRemoteRegistryExtension(ctx, "extensionID", extensionID)
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

var mockLocalRegistryExtensionIDPrefix **string

// GetLocalRegistryExtensionIDPrefix returns the extension ID prefix (if any) of extensions in the
// local registry.
func GetLocalRegistryExtensionIDPrefix() (*string, error) {
	if mockLocalRegistryExtensionIDPrefix != nil {
		return *mockLocalRegistryExtensionIDPrefix, nil
	}
	if envvar.SourcegraphDotComMode() {
		return nil, nil
	}
	name, err := GetLocalRegistryName()
	if err != nil {
		return nil, nil
	}
	return &name, nil
}

// getRemoteRegistryURL returns the remote registry URL from site configuration, or nil if there is
// none. If an error exists while parsing the value in site configuration, the error is returned.
func getRemoteRegistryURL() (*url.URL, error) {
	pc := conf.Platform()
	if pc == nil || pc.RemoteRegistryURL == "" {
		return nil, nil
	}
	return url.Parse(pc.RemoteRegistryURL)
}

var mockGetRemoteRegistryExtension func(field, value string) (*registry.Extension, error)

// GetRemoteRegistryExtension gets the remote registry extension and rewrites its fields to be from
// the frame-of-reference of this site. The field is either "uuid" or "extensionID".
func GetRemoteRegistryExtension(ctx context.Context, field, value string) (*registry.Extension, error) {
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
	return x, err
}

// ListRemoteRegistryExtensions lists the remote registry extensions and rewrites their fields to be
// from the frame-of-reference of this site.
func ListRemoteRegistryExtensions(ctx context.Context, query string) ([]*registry.Extension, error) {
	registryURL, err := getRemoteRegistryURL()
	if registryURL == nil || err != nil {
		return nil, err
	}

	xs, err := registry.List(ctx, registryURL, query)
	if err != nil {
		return nil, err
	}
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
