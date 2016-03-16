package httpapi

import (
	"net/http"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoTreeGetOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx, _ := handlerutil.Client(r)
	tc, _, _, err := handlerutil.GetTreeEntryCommon(ctx, mux.Vars(r), &opt)
	if err != nil {
		return err
	}
	return writeJSON(w, tc.Entry)
}

func serveRepoTreeList(w http.ResponseWriter, r *http.Request) error {
	repoRev, err := sourcegraph.UnmarshalRepoRevSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	ctx, cl := handlerutil.Client(r)
	treeList, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: repoRev})
	if err != nil {
		return err
	}
	return writeJSON(w, treeList)
}
