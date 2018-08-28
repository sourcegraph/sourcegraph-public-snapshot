package app

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// serveRegistryExtensionBundle serves the bundled JavaScript source file for an extension in the
// registry as a raw JavaScript file.
func serveRegistryExtensionBundle(w http.ResponseWriter, r *http.Request) {
	if conf.Platform() == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	releaseIDStr := strings.TrimSuffix(mux.Vars(r)["RegistryExtensionReleaseID"], ".js")
	releaseID, err := strconv.ParseInt(releaseIDStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bundle, err := db.RegistryExtensionReleases.GetBundle(r.Context(), releaseID)
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
	w.Header().Del("Access-Control-Allow-Credentials")
	w.Header().Del("Access-Control-Allow-Origin")

	// We want to cache forever because an extension release is immutable, except that if the
	// database is reset and and the registry_extension_releases.id sequence starts over, we don't
	// want stale data from pre-reset. So, assume that the presence of a query string means that it
	// includes some identifier that changes when the database is reset.
	if r.URL.RawQuery != "" {
		w.Header().Set("Cache-Control", "max-age=604800, private, immutable")
	}
	w.Write(bundle)
}
