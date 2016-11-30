package httpapi

import (
	"errors"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/cznic/mathutil"
	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/honey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

type treeEntry struct {
	sourcegraph.TreeEntry
	IncludedAnnotations *sourcegraph.AnnotationList
}

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	if actor := auth.ActorFromContext(r.Context()); actor != nil {
		ev := honey.Event("repo_tree")
		ev.AddField("uid", actor.UID)
		ev.AddField("login", actor.Login)
		ev.AddField("email", actor.Email)
		ev.Send()
	}

	vars := mux.Vars(r)
	orig := routevar.ToTreeEntry(vars)
	repoRev, err := resolveLocalRepoRev(r.Context(), orig.RepoRev)
	if err != nil {
		return err
	}

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
			Entry: entrySpec,
			Range: &opt.FileRange,
		})
		if err == nil {
			res.IncludedAnnotations = anns
		} else {
			log15.Warn("Error optimistically including annotations in serveRepoTree", "entry", entrySpec, "err", err)
		}
	}

	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, 60*time.Second); clientCached || err != nil {
		return err
	}

	return writeJSON(w, res)
}
