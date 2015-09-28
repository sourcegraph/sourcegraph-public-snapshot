package app

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// serveRepoChangeset handles routes for both RepoChangeset and RepoChangesetChanges.
func serveRepoChangeset(w http.ResponseWriter, r *http.Request) error {
	// TODO(gbbr): Construct payload only in UI and use that method here.
	// The purpose of the 'ui' package is the construct payloads for UI components
	// and the below code is doing the same thing (not DRY) making the ui package
	// redundant. This code is more or less a duplicate of its counter-part:
	// ui/repo_changeset.go
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}
	if !repoEnabledFrameChangesets(rc.Repo) {
		return fmt.Errorf("not a valid app")
	}
	// Retrieve Changeset
	changesetSpec := &sourcegraph.ChangesetSpec{
		Repo: vc.RepoRevSpec.RepoSpec,
		ID:   id,
	}
	cs, err := cl.Changesets.Get(ctx, changesetSpec)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return &handlerutil.HTTPErr{Status: http.StatusNotFound, Err: errors.New("changeset does not exist")}
		}
		return err
	}
	ds := cs.DeltaSpec
	// Compute delta (actual merge-base, commit IDs and build status for both revs)
	delta, err := cl.Deltas.Get(ctx, &sourcegraph.DeltaSpec{
		Base: ds.Base,
		Head: ds.Head,
	})
	if err != nil {
		// if any revision in this changeset does not exist, alert the user.
		// This may happen after branches are deleted.
		if grpc.Code(err) == codes.NotFound {
			return displayExpiredChangesetTemplate(w, r, rc, vc, cs, ds)
		}
		return err
	}
	reviews, err := cl.Changesets.ListReviews(ctx, &sourcegraph.ChangesetListReviewsOp{
		Repo:        vc.RepoRevSpec.RepoSpec,
		ChangesetID: cs.ID,
	})
	if err != nil {
		return err
	}
	// Compute the tip of the base branch
	baseTip, err := cl.Repos.GetCommit(ctx, &ds.Base)
	if err != nil {
		return err
	}
	// Retrieve commit list
	commitList, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: ds.Base.RepoSpec,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Base:        string(delta.BaseCommit.ID),
			Head:        string(delta.HeadCommit.ID),
			ListOptions: sourcegraph.ListOptions{PerPage: -1},
		},
	})
	if err != nil {
		return err
	}
	if len(commitList.Commits) == 0 {
		// if there are no commits between these revisions, this changeset
		// has already been merged.
		return displayExpiredChangesetTemplate(w, r, rc, vc, cs, ds)
	}
	sort.Sort(byDate(commitList.Commits))
	// If this route is for changes, request the diffs too
	var (
		files  *sourcegraph.DeltaFiles
		filter string
	)
	if f, ok := v["Filter"]; ok {
		filter = f
	}
	opt := sourcegraph.DeltaListFilesOptions{
		Formatted: false,
		Tokenized: true,
		Filter:    filter,
	}
	files, err = cl.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
		Ds:  *ds,
		Opt: &opt,
	})
	if err != nil {
		return err
	}
	// Augment commits with data from People
	augmentedCommits := make([]*payloads.AugmentedCommit, len(commitList.Commits))
	for i, c := range commitList.Commits {
		augmentedCommits[i], err = handlerutil.AugmentCommit(r, delta.HeadRepo.URI, c)
		if err != nil {
			return err
		}
	}
	events, err := cl.Changesets.ListEvents(ctx, changesetSpec)
	if err != nil {
		return err
	}
	return tmpl.Exec(r, w, "repo/changeset.html", http.StatusOK, nil, &struct {
		tmpl.Common `json:"-"`
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		payloads.Changeset

		FileFilter string
	}{
		Common:        tmpl.Common{FullWidth: true},
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		FileFilter:    filter,

		Changeset: payloads.Changeset{
			Changeset: cs,
			Delta:     delta,
			BaseTip:   baseTip,
			Commits:   augmentedCommits,
			Files:     files,
			Reviews:   reviews.Reviews,
			Events:    events.Events,
		},
	})
}

func displayExpiredChangesetTemplate(w http.ResponseWriter, r *http.Request, rc *handlerutil.RepoCommon, vc *handlerutil.RepoRevCommon, cs *sourcegraph.Changeset, ds *sourcegraph.DeltaSpec) error {
	return tmpl.Exec(r, w, "repo/changeset.notfound.html", http.StatusOK, nil, &struct {
		tmpl.Common `json:"-"`
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		Changeset *sourcegraph.Changeset
		DeltaSpec *sourcegraph.DeltaSpec
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		Changeset:     cs,
		DeltaSpec:     ds,
	})
}

// serveRepoChangesetList serves a list of changesets for the current repository.
func serveRepoChangesetList(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := httpctx.FromRequest(r), handlerutil.APIClient(r)
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}
	if !repoEnabledFrameChangesets(rc.Repo) {
		return fmt.Errorf("not a valid app")
	}
	op := sourcegraph.ChangesetListOp{Repo: mux.Vars(r)["Repo"]}
	var param struct{ Closed bool }
	schemautil.Decode(&param, r.URL.Query())
	op.Closed = param.Closed
	op.Open = !param.Closed
	op.ListOptions.PerPage = math.MaxInt32 // TODO(sqs): properly impl pagination - remove sort.Sort here, put in backend
	list, err := cl.Changesets.List(ctx, &op)
	if err != nil {
		return err
	}
	sort.Sort(changesetsByDate(list.Changesets))
	return tmpl.Exec(r, w, "repo/changeset.list.html", http.StatusOK, nil, &struct {
		tmpl.Common `json:"-"`
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		List []*sourcegraph.Changeset
		Op   sourcegraph.ChangesetListOp
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		List:          list.Changesets,
		Op:            op,
	})
}

// byDate implements the sorting interface for sorting
// a list of commits by authorship date.
type byDate []*vcs.Commit

func (b byDate) Len() int { return len(b) }

func (b byDate) Less(i, j int) bool {
	return b[i].Author.Date.Time().Before(b[j].Author.Date.Time())
}

func (b byDate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

type changesetsByDate []*sourcegraph.Changeset

func (cs changesetsByDate) Len() int { return len(cs) }

func (cs changesetsByDate) Less(i, j int) bool {
	return cs[i].CreatedAt.Time().After(cs[j].CreatedAt.Time())
}

func (cs changesetsByDate) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
