package httpapi

import (
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveAnnotations(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	// HACK: Make the Entry.RepoRev.Repo value available at the
	// keypath Repo so that we can decode it into the Repo field. You
	// can't specify a dotted keypath in the url struct tag in
	// gorilla/schema; this is a workaround.
	q := r.URL.Query()
	q["Repo"] = q["Entry.RepoRev.Repo"]
	delete(q, "Entry.RepoRev.Repo")
	var tmp struct {
		Repo repoIDOrPath
		sourcegraph.AnnotationsListOptions
	}
	if err := schemaDecoder.Decode(&tmp, q); err != nil {
		return err
	}
	opt := tmp.AnnotationsListOptions
	if tmp.Repo != "" {
		var err error
		opt.Entry.RepoRev.Repo, err = getRepoID(ctx, tmp.Repo)
		if err != nil {
			return err
		}
	}

	anns, err := cl.Annotations.List(ctx, &opt)
	if err != nil {
		return err
	}
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}
	return writeJSON(w, anns)
}
