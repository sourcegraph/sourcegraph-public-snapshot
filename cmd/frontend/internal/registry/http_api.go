package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
)

const (
	// APIVersion is a string that uniquely identifies this API version.
	APIVersion = "20180621"

	// AcceptHeader is the value of the "Accept" HTTP request header sent by the client.
	AcceptHeader = "application/vnd.sourcegraph.api+json;version=" + APIVersion

	// MediaTypeHeaderName is the name of the HTTP response header whose value the client expects to
	// equal MediaType.
	MediaTypeHeaderName = "X-Sourcegraph-Media-Type"

	// MediaType is the client's expected value for the MediaTypeHeaderName HTTP response header.
	MediaType = "sourcegraph.v" + APIVersion + "; format=json"
)

// HandleRegistry serves the external HTTP API for the extension registry for Sourcegraph.com
// only. All other extension registries have been removed. See
// https://docs.google.com/document/d/10vtoe-kpNvVZ8Etrx34bSCoTaCCHxX8o3ncCmuErPZo/edit.
func HandleRegistry(w http.ResponseWriter, r *http.Request) (err error) {
	if !dotcom.SourcegraphDotComMode() {
		http.Error(w, "no local extension registry exists", http.StatusNotFound)
		return nil
	}

	// Identify this response as coming from the registry API.
	w.Header().Set(MediaTypeHeaderName, MediaType)

	// The response differs based on some request headers, and we need to tell caches which ones.
	//
	// Accept, User-Agent: because these encode the registry client's API version, and responses are
	// not cacheable across versions.
	w.Header().Set("Vary", "Accept, User-Agent")

	// Validate API version.
	if v := r.Header.Get("Accept"); v != AcceptHeader {
		http.Error(w, fmt.Sprintf("invalid Accept header: expected %q", AcceptHeader), http.StatusBadRequest)
		return nil
	}

	urlPath := strings.TrimPrefix(r.URL.Path, "/.api")

	const extensionsPath = "/registry/extensions"
	var result any
	switch {
	case urlPath == extensionsPath:
		result = filterRegistryExtensions(getFrozenRegistryData(), r.URL.Query().Get("q"))

	case urlPath == extensionsPath+"/featured":
		result = []struct{}{}

	case strings.HasPrefix(urlPath, extensionsPath+"/"):
		var (
			spec = strings.TrimPrefix(urlPath, extensionsPath+"/")
			x    *Extension
		)
		switch {
		case strings.HasPrefix(spec, "uuid/"):
			x = findRegistryExtension(getFrozenRegistryData(), "uuid", strings.TrimPrefix(spec, "uuid/"))
		case strings.HasPrefix(spec, "extension-id/"):
			x = findRegistryExtension(getFrozenRegistryData(), "extensionID", strings.TrimPrefix(spec, "extension-id/"))
		default:
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		if x == nil {
			w.Header().Set("Cache-Control", "max-age=5, private")
			http.Error(w, "extension not found", http.StatusNotFound)
			return nil
		}
		result = x

	default:
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	w.Header().Set("Cache-Control", "max-age=120, private")
	return json.NewEncoder(w).Encode(result)
}

// filterRegistryExtensions returns the subset of extensions that match the query. It does not
// modify its arguments.
func filterRegistryExtensions(extensions []*Extension, query string) []*Extension {
	if query == "" {
		return extensions
	}

	query = strings.ToLower(query)
	var keep []*Extension
	for _, x := range extensions {
		if strings.Contains(strings.ToLower(x.ExtensionID), query) {
			keep = append(keep, x)
		}
	}
	return keep
}

// findRegistryExtension returns the first (and, hopefully, only, although that's not enforced)
// extension whose field matches the given value, or nil if none match.
func findRegistryExtension(extensions []*Extension, field, value string) *Extension {
	match := func(x *Extension) bool {
		switch field {
		case "uuid":
			return x.UUID == value
		case "extensionID":
			return x.ExtensionID == value
		default:
			panic("unexpected field: " + field)
		}
	}

	for _, x := range extensions {
		if match(x) {
			return x
		}
	}
	return nil
}
