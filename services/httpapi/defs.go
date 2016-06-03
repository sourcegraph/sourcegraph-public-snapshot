package httpapi

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveDef(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	var opt sourcegraph.DefGetOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	def, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), &opt)
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

func resolveDef(ctx context.Context, def routevar.DefAtRev) (*sourcegraph.DefSpec, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: sourcegraph.RepoSpec{URI: def.Repo}, Rev: def.Rev})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.DefSpec{
		Repo:     def.Repo,
		CommitID: res.CommitID,
		UnitType: def.UnitType,
		Unit:     def.Unit,
		Path:     def.Path,
	}, nil
}
