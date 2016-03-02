package app

import (
	"net/http"
	"sort"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/pkg/vcs"

	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveRepoCommit(w http.ResponseWriter, r *http.Request) error {
	ctx, cl, _, err := handlerutil.RepoClient(r)
	if err != nil {
		return err
	}

	rc, vc, err := handlerutil.GetRepoAndRevCommon(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	headRevSpec := vc.RepoRevSpec

	// The first commit on a branch has no base. Handle this case.
	var baseRevSpec *sourcegraph.RepoRevSpec
	if len(vc.RepoCommit.Parents) > 0 {
		baseCommitID := string(vc.RepoCommit.Parents[0])
		baseRevSpec = &sourcegraph.RepoRevSpec{
			RepoSpec: rc.Repo.RepoSpec(),
			Rev:      baseCommitID,
			CommitID: baseCommitID,
		}
	}

	var delta *sourcegraph.Delta
	var deltaFiles *sourcegraph.DeltaFiles
	if baseRevSpec != nil {
		ds := sourcegraph.DeltaSpec{Base: *baseRevSpec, Head: headRevSpec}

		delta, err = cl.Deltas.Get(ctx, &ds)
		if err != nil {
			return err
		}

		var err error
		deltaFiles, err = cl.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
			Ds:  ds,
			Opt: &sourcegraph.DeltaListFilesOptions{Filter: r.URL.Query().Get("filter")},
		})
		if err != nil {
			return err
		}
	}

	tmplData := struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		Delta      *sourcegraph.Delta
		DeltaFiles *sourcegraph.DeltaFiles
		DeltaSpec  sourcegraph.DeltaSpec

		// TODO(beyang): additional hacks like the one above
		ShowFiles bool
		Filter    string

		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		Delta:         delta,
		DeltaFiles:    deltaFiles,

		ShowFiles: true,
	}
	if delta != nil {
		tmplData.DeltaSpec = delta.DeltaSpec()
	}

	return tmpl.Exec(r, w, "repo/commit.html", http.StatusOK, nil, &tmplData)
}

func serveRepoCommits(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListCommitsOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx, cl, _, err := handlerutil.RepoClient(r)
	if err != nil {
		return err
	}

	rc, vc, err := handlerutil.GetRepoAndRevCommon(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	// Set defaults for ListCommits call options.
	listCommitsOpt := opt
	if listCommitsOpt.Head == "" {
		listCommitsOpt.Head = vc.RepoRevSpec.CommitID
	}
	if listCommitsOpt.PerPage == 0 {
		listCommitsOpt.PerPage = 10
	}
	commits0, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{Repo: rc.Repo.RepoSpec(), Opt: &listCommitsOpt})
	if err != nil {
		return err
	}

	pg, err := paginatePrevNext(opt, commits0.StreamResponse)
	if err != nil {
		return err
	}

	commitDays, err := handlerutil.AugmentAndGroupCommitsByDay(ctx, commits0.Commits, rc.Repo.URI)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/commits.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		CommitDays []*handlerutil.DayOfAugmentedCommits
		PageLinks  []pageLink
		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		CommitDays:    commitDays,
		PageLinks:     pg,
	})
}

func serveRepoBranches(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListBranchesOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx, cl, _, err := handlerutil.RepoClient(r)
	if err != nil {
		return err
	}

	rc, err := handlerutil.GetRepoCommon(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	opt.IncludeCommit = true

	branches, err := cl.Repos.ListBranches(ctx, &sourcegraph.ReposListBranchesOp{Repo: rc.Repo.RepoSpec(), Opt: &opt})
	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(vcs.ByAuthorDate(branches.Branches)))

	// TODO(x): paginate

	return tmpl.Exec(r, w, "repo/branches.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Branches          []*vcs.Branch
		BehindAheadBranch string
		PageLinks         []pageLink
		tmpl.Common
	}{
		RepoCommon:        *rc,
		Branches:          branches.Branches,
		BehindAheadBranch: opt.BehindAheadBranch,
	})
}

func serveRepoTags(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListTagsOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx, cl, _, err := handlerutil.RepoClient(r)
	if err != nil {
		return err
	}

	rc, err := handlerutil.GetRepoCommon(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	tags, err := cl.Repos.ListTags(ctx, &sourcegraph.ReposListTagsOp{Repo: rc.Repo.RepoSpec(), Opt: &opt})
	if err != nil {
		return err
	}

	// TODO(x): paginate

	return tmpl.Exec(r, w, "repo/tags.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Tags      []*vcs.Tag
		PageLinks []pageLink
		tmpl.Common
	}{
		RepoCommon: *rc,
		Tags:       tags.Tags,
	})
}
