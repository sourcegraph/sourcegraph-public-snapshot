package httpapi

import (
	"net/http"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	vars := mux.Vars(r)
	repoRev, err := sourcegraph.UnmarshalRepoRevSpec(vars)
	if err != nil {
		return err
	}

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: repoRev,
		Path:    path.Clean(strings.TrimPrefix(vars["Path"], "/")),
	}

	resolvedRev, _, err := handlerutil.ResolveSrclibDataVersion(ctx, entrySpec)
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	}
	if resolvedRev.CommitID != "" {
		entrySpec.RepoRev.CommitID = resolvedRev.CommitID
	}

	var opt sourcegraph.RepoTreeGetOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, entry)
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
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, treeList)
}

func serveRepoTreeSearch(w http.ResponseWriter, r *http.Request) error {
	repoRev, err := sourcegraph.UnmarshalRepoRevSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	var opt sourcegraph.RepoTreeSearchOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	ctx, cl := handlerutil.Client(r)
	treeSearch, err := cl.RepoTree.Search(ctx, &sourcegraph.RepoTreeSearchOp{
		Rev: repoRev,
		Opt: &opt,
	})
	if err != nil {
		return err
	}
	return writeJSON(w, treeSearch)
}
