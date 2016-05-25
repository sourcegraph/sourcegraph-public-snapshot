package httpapi

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/cznic/mathutil"
	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

type treeEntry struct {
	sourcegraph.TreeEntry
	IncludedAnnotations *sourcegraph.AnnotationList
}

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	vars := mux.Vars(r)
	origRepoRev := routevar.ToRepoRev(vars)
	repoRev, err := resolveRepoRev(ctx, origRepoRev)
	if err != nil {
		return err
	}

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    path.Clean(strings.TrimPrefix(vars["Path"], "/")),
	}

	var opt sourcegraph.RepoTreeGetOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
	if err != nil {
		return err
	}

	// Limit the size of files we return to prevent certain parts of the
	// UI (like DefInfo.js) from becoming unresponsive in the browser.
	//
	// TODO support displaying file ranges on the front-end instead.
	const maxFileSize = 512 * 1024
	size := mathutil.Max(len(entry.Contents), len(entry.ContentsString))
	if entry.Type == sourcegraph.FileEntry && size > maxFileSize {
		return &errcode.HTTPErr{
			Status: http.StatusRequestEntityTooLarge,
			Err:    errors.New("file too large (size=" + string(size) + ")"),
		}
	}

	res := treeEntry{TreeEntry: *entry}

	// As an optimization, optimistically include the file's
	// annotations (if this entry is a file), to save a round-trip in
	// most cases. Don't do this if the file is large; currently
	// the heuristic is ~ 2500 lines at avg. 40 chars per line
	if entry.Type == sourcegraph.FileEntry && len(entry.ContentsString) < (40*2500) {
		anns, err := cl.Annotations.List(ctx, &sourcegraph.AnnotationsListOptions{
			Entry: entrySpec,
			Range: &opt.FileRange,
		})
		if err == nil {
			res.IncludedAnnotations = anns
		} else {
			log15.Warn("Error optimistically including annotations in serveRepoTree", "entry", entrySpec, "err", err)
		}
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, res)
}

func serveRepoTreeList(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoRev, err := resolveRepoRev(ctx, routevar.ToRepoRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	treeList, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: *repoRev})
	if err != nil {
		return err
	}
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, treeList)
}

func serveRepoTreeSearch(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoRev, err := resolveRepoRev(ctx, routevar.ToRepoRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	var opt sourcegraph.RepoTreeSearchOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	treeSearch, err := cl.RepoTree.Search(ctx, &sourcegraph.RepoTreeSearchOp{
		Rev: *repoRev,
		Opt: &opt,
	})
	if err != nil {
		return err
	}
	return writeJSON(w, treeSearch)
}
