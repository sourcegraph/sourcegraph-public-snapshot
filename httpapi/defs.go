package httpapi

import (
	"net/http"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveDef(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.DefGetOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	defSpec, err := getDefSpec(r)
	if err != nil {
		return err
	}

	def, err := s.Defs.Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: &opt})
	if err != nil {
		return err
	}
	return writeJSON(w, def)
}

func getDefSpec(r *http.Request) (defSpec sourcegraph.DefSpec, err error) {
	v := mux.Vars(r)
	s := handlerutil.APIClient(r)

	_, repoRevSpec, _, err := handlerutil.GetRepoAndRev(r, s.Repos)
	if err != nil {
		return sourcegraph.DefSpec{}, err
	}

	spec := sourcegraph.DefSpec{
		Repo:     repoRevSpec.URI,
		CommitID: repoRevSpec.CommitID,
		UnitType: v["UnitType"],
		Unit:     v["Unit"],
		Path:     v["Path"],
	}
	if spec.Repo == "" {
		panic("empty RepoURI")
	}
	if spec.UnitType == "" {
		panic("empty UnitType")
	}
	return spec, nil
}

func serveDefs(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

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

	defs, err := s.Defs.List(ctx, &opt)
	if err != nil {
		return err
	}
	return writeJSON(w, defs)
}
