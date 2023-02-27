package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
)

func init() {
	if envvar.SourcegraphDotComMode() {
		frontendregistry.HandleRegistry = handleRegistry
	}
}

// handleRegistry serves the external HTTP API for the extension registry for Sourcegraph.com
// only. All other extension registries have been removed. See
// https://docs.google.com/document/d/10vtoe-kpNvVZ8Etrx34bSCoTaCCHxX8o3ncCmuErPZo/edit.
func handleRegistry(w http.ResponseWriter, r *http.Request) (err error) {
	// Identify this response as coming from the registry API.
	w.Header().Set(registry.MediaTypeHeaderName, registry.MediaType)

	// The response differs based on some request headers, and we need to tell caches which ones.
	//
	// Accept, User-Agent: because these encode the registry client's API version, and responses are
	// not cacheable across versions.
	w.Header().Set("Vary", "Accept, User-Agent")

	// Validate API version.
	if v := r.Header.Get("Accept"); v != registry.AcceptHeader {
		http.Error(w, fmt.Sprintf("invalid Accept header: expected %q", registry.AcceptHeader), http.StatusBadRequest)
		return nil
	}

	urlPath := strings.TrimPrefix(r.URL.Path, "/.api")

	const extensionsPath = "/registry/extensions"
	var result any
	switch {
	case urlPath == extensionsPath:
		result = frontendregistry.FilterRegistryExtensions(getFrozenRegistryData(), r.URL.Query().Get("q"))

	case urlPath == extensionsPath+"/featured":
		result = []struct{}{}

	case strings.HasPrefix(urlPath, extensionsPath+"/"):
		var (
			spec = strings.TrimPrefix(urlPath, extensionsPath+"/")
			x    *registry.Extension
		)
		switch {
		case strings.HasPrefix(spec, "uuid/"):
			x = frontendregistry.FindRegistryExtension(getFrozenRegistryData(), "uuid", strings.TrimPrefix(spec, "uuid/"))
		case strings.HasPrefix(spec, "extension-id/"):
			x = frontendregistry.FindRegistryExtension(getFrozenRegistryData(), "extensionID", strings.TrimPrefix(spec, "extension-id/"))
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
