package httpapi

import (
	"errors"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/cznic/mathutil"
	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

type treeEntry struct {
	sourcegraph.TreeEntry
	IncludedAnnotations *sourcegraph.AnnotationList
}

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	orig := routevar.ToTreeEntry(vars)
	repoRev, err := resolveLocalRepoRev(r.Context(), orig.RepoRev)
	if err != nil {
		return err
	}

	// Optimization: Ensure our language server is ready to start
	// responding to requests
	go langp.DefaultClient.Prepare(r.Context(), &langp.RepoRev{
		Repo:   orig.RepoRev.Repo,
		Commit: repoRev.CommitID,
	})

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    orig.Path,
	}

	var opt sourcegraph.RepoTreeGetOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	entry, err := backend.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
	if err != nil {
		return err
	}

	// Limit the size of files we return to prevent certain parts of the
	// UI (like DefInfo.js) from becoming unresponsive in the browser.
	//
	// TODO support displaying file ranges on the front-end instead.
	const maxFileSize = 5 * 1024 * 1024
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
		anns, err := backend.Annotations.List(r.Context(), &sourcegraph.AnnotationsListOptions{
			Entry:        entrySpec,
			Range:        &opt.FileRange,
			NoSrclibAnns: opt.NoSrclibAnns,
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
	unresolvedRepoRev := routevar.ToRepoRev(mux.Vars(r))
	repoRev, err := resolveLocalRepoRev(r.Context(), unresolvedRepoRev)
	if err != nil {
		return err
	}

	// Optimization: Ensure our language server is ready to start
	// responding to requests
	go langp.DefaultClient.Prepare(r.Context(), &langp.RepoRev{
		Repo:   unresolvedRepoRev.Repo,
		Commit: repoRev.CommitID,
	})

	treeList, err := backend.RepoTree.List(r.Context(), &sourcegraph.RepoTreeListOp{Rev: *repoRev})
	if err != nil {
		return err
	}
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, treeList)
}
