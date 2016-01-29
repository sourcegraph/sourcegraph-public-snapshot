package app

import (
	"net/http"
	"sort"

	"github.com/rogpeppe/rog-go/parallel"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"

	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoCommit(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r)
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
	var files *sourcegraph.DeltaFiles
	if baseRevSpec != nil {
		ds := sourcegraph.DeltaSpec{Base: *baseRevSpec, Head: headRevSpec}

		delta, err = cl.Deltas.Get(ctx, &ds)
		if err != nil {
			return err
		}

		par := parallel.NewRun(3)
		par.Do(func() (err error) {
			opt := sourcegraph.DeltasListFilesOp{
				Ds: ds,
				Opt: &sourcegraph.DeltaListFilesOptions{
					Formatted: false,
					Tokenized: true,
					Filter:    r.URL.Query().Get("filter"),
				},
			}
			files, err = cl.Deltas.ListFiles(ctx, &opt)
			return
		})
		if err := par.Wait(); err != nil {
			return err
		}
	}

	tmplData := struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		Delta     *sourcegraph.Delta
		DiffData  *sourcegraph.DeltaFiles
		DeltaSpec sourcegraph.DeltaSpec

		DeltaListDefsOpt *sourcegraph.DeltaListDefsOptions

		// TODO(beyang): additional hacks like the one above
		ShowFiles     bool
		Filter        string
		OverThreshold bool

		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		Delta:         delta,
		DiffData:      files,

		ShowFiles: true,
	}
	if delta != nil {
		tmplData.DeltaSpec = delta.DeltaSpec()
	}
	if files != nil {
		tmplData.OverThreshold = files.OverThreshold
	}

	return tmpl.Exec(r, w, "repo/commit.html", http.StatusOK, nil, &tmplData)
}

func serveRepoCommits(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoListCommitsOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r)
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
	commits0, err := apiclient.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{Repo: rc.Repo.RepoSpec(), Opt: &listCommitsOpt})
	if err != nil {
		return err
	}

	pg, err := paginatePrevNext(opt, commits0.StreamResponse)
	if err != nil {
		return err
	}

	commitDays, err := handlerutil.AugmentAndGroupCommitsByDay(r, commits0.Commits, rc.Repo.URI)
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

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	opt.IncludeCommit = true

	branches, err := apiclient.Repos.ListBranches(ctx, &sourcegraph.ReposListBranchesOp{Repo: rc.Repo.RepoSpec(), Opt: &opt})
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

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	tags, err := apiclient.Repos.ListTags(ctx, &sourcegraph.ReposListTagsOp{Repo: rc.Repo.RepoSpec(), Opt: &opt})
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
