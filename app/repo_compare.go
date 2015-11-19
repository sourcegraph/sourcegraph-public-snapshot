package app

import (
	"errors"
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"

	"github.com/sourcegraph/mux"
)

// serveRepoCompare handles both routes RepoCompare and RepoCompareAll and
// differentiates by checking the suffix in r.URL
func serveRepoCompare(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}
	v := mux.Vars(r)
	if v["Head"] == "" {
		return errors.New("must specify head branch")
	}
	repoSpec := vc.RepoRevSpec.RepoSpec
	head := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, Rev: v["Head"]}

	ds := sourcegraph.DeltaSpec{Base: vc.RepoRevSpec, Head: head}
	opt := sourcegraph.DeltaListFilesOptions{
		Formatted: false,
		Tokenized: true,
		Filter:    r.URL.Query().Get("filter"),
	}

	files, err := cl.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{Ds: ds, Opt: &opt})
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/compare.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		DiffData      *sourcegraph.DeltaFiles
		DeltaSpec     sourcegraph.DeltaSpec
		tmpl.Common   `json:"-"`
		OverThreshold bool
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		DiffData:      files,
		DeltaSpec:     ds,
		OverThreshold: files.OverThreshold,
	})
}
