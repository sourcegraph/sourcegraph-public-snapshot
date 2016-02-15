package app

import (
	"errors"
	"net/http"

	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"

	"github.com/sourcegraph/mux"
)

// serveRepoCompare handles both routes RepoCompare and RepoCompareAll and
// differentiates by checking the suffix in r.URL
func serveRepoCompare(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(ctx, mux.Vars(r))
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
	files, err := cl.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
		Ds:  ds,
		Opt: &sourcegraph.DeltaListFilesOptions{Filter: r.URL.Query().Get("filter")},
	})
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
		OverThreshold: diffSizeIsOverThreshold(files.DiffStat()),
	})
}
