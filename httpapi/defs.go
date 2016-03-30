package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveDef(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	var opt sourcegraph.DefGetOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	def, _, _, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), &opt)
	if err != nil {
		return err
	}
	return writeJSON(w, def)
}

func serveDefs(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	// Caching
	lastMod, err := getLastModForRepoRevs(r, opt.RepoRevs)
	if err != nil {
		return err
	}
	if clientCached, err := writeCacheHeaders(w, r, lastMod, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	defs, err := cl.Defs.List(ctx, &opt)
	if err != nil {
		return err
	}
	return writeJSON(w, defs)
}
