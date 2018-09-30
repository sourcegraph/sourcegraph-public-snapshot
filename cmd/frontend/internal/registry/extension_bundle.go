package registry

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func init() {
	frontendregistry.HandleRegistryExtensionBundle = handleRegistryExtensionBundle
}

// handleRegistryExtensionBundle serves the bundled JavaScript source file or the source map for an
// extension in the registry as a raw JavaScript or JSON file.
func handleRegistryExtensionBundle(w http.ResponseWriter, r *http.Request) {
	if conf.Extensions() == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	filename := mux.Vars(r)["RegistryExtensionReleaseFilename"]
	ext := filepath.Ext(filename)
	wantSourceMap := ext == ".map"
	releaseIDStr := strings.TrimSuffix(filename, ext)
	releaseID, err := strconv.ParseInt(releaseIDStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bundle, sourceMap, err := dbReleases{}.GetArtifacts(r.Context(), releaseID)
	if errcode.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 🚨 SECURITY: Prevent this URL from being used in a <script> tag from other sites, because
	// hosting user-provided scripts on this domain would let attackers steal sensitive data from
	// anyone they lure to the attacker's site.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("X-Frame-Options", "deny")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Allow downstream Sourcegraph sites' clients to access this file directly.
	w.Header().Del("Access-Control-Allow-Credentials") // credentials are not needed
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// We want to cache forever because an extension release is immutable, except that if the
	// database is reset and and the registry_extension_releases.id sequence starts over, we don't
	// want stale data from pre-reset. So, assume that the presence of a query string means that it
	// includes some identifier that changes when the database is reset.
	if r.URL.RawQuery != "" {
		w.Header().Set("Cache-Control", "max-age=604800, private, immutable")
	}
	var data []byte
	if wantSourceMap {
		if sourceMap == nil {
			http.Error(w, "extension has no source map", http.StatusNotFound)
			return
		}
		data = sourceMap
	} else {
		data = bundle
	}
	w.Write(data)

	if !wantSourceMap && sourceMap != nil {
		// Append `//# sourceMappingURL=` directive to JS bundle if we have a source map. It is
		// necessary to provide the absolute URL because the JS bundle is not loaded directly (e.g.,
		// via importScripts); it is saved to a blob URL and then executed, which means any relative
		// source map URL would be interpreted relative to the blob URL (so a relative URL wouldn't
		// work). Also, we can't rely on the original sourceMappingURL directive (if provided at
		// publish time) because it has no way of knowing the absolute URL to the source map.
		//
		// This implementation is not ideal because it means the JS bundle's contents depend on the
		// app URL, which makes it technically not immutable. But given the blob URL constraint
		// mentioned above, it's the best known solution.
		if appURL, _ := url.Parse(conf.Get().AppURL); appURL != nil {
			sourceMapURL := appURL.ResolveReference(&url.URL{Path: path.Join(path.Dir(r.URL.Path), releaseIDStr+".map")}).String()
			fmt.Fprintf(w, "\n//# sourceMappingURL=%s", sourceMapURL)
		}
	}
}
