// Package wellknown provides HTTP handlers that live at "well-known
// URIs" and describe site-wide configuration. When imported, as a
// side effect it adds its registration function to cli.ServeMuxFuncs,
// so it will be mounted on the sgx package's HTTP server.
//
// So-called "Well-Known URIs" enable discovery of site-wide
// configuration data for an HTTP server. See [RFC
// 5785](https://tools.ietf.org/html/rfc5785) for more information
// about Well-Known URIs.
package wellknown

import (
	"encoding/json"
	"log"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	cli.ServeMuxFuncs = append(cli.ServeMuxFuncs, AddConfigHandler)
}

// ConfigPath is the well-known URI path for the Sourcegraph host's
// configuration data. Typically callers should just use
// AddConfigHandler to set up the handler at this path.
const ConfigPath = "/.well-known/sourcegraph"

// AddConfigHandler adds a HTTP handler at the path
// "/.well-known/sourcegraph" that describes the configuration for the
// Sourcegraph host. The configuration is obtained by calling the
// Meta.Config API method.
func AddConfigHandler(mux *http.ServeMux) {
	mux.HandleFunc(ConfigPath, func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetRouteName(r, "well-known")
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		ctx := httpctx.FromRequest(r)
		cl, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		config, err := cl.Meta.Config(ctx, &pbtypes.Void{})
		if err != nil {
			log.Printf("Error serving %s: %s.", r.URL, err)
			// Can't use package errcode due to import cycle.
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.Header().Set("cache-control", "private, max-age=300")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
}
