package app

import (
	"net/http"

	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"

	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// defaultBuildListOptions takes the provided BuildListOptions, and returns a copy with
// missing fields to their default values.
func defaultBuildListOptions(opt sourcegraph.BuildListOptions) sourcegraph.BuildListOptions {
	if opt.PerPage == 0 {
		opt.PerPage = 50
	}
	if opt.Sort == "" {
		opt.Sort = "created_at"
		opt.Direction = "desc"
	}
	return opt
}

// TODO(shurcooL): Find out where it's used, if it's still needed, and whether pagination support should be added here.
func serveBuilds(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.BuildListOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	opt = defaultBuildListOptions(opt)
	builds, err := apiclient.Builds.List(ctx, &opt)
	if err != nil {
		return err
	}

	type tab struct {
		Name string
		sourcegraph.BuildListOptions
	}
	tabs := []tab{
		{"All", sourcegraph.BuildListOptions{Sort: "bid", Direction: "desc"}},
		{"Priority Queue", sourcegraph.BuildListOptions{Queued: true, Sort: "priority", Direction: "desc"}},
		{"Active", sourcegraph.BuildListOptions{Active: true, Sort: "updated_at", Direction: "desc"}},
		{"Ended", sourcegraph.BuildListOptions{Ended: true, Sort: "updated_at", Direction: "desc"}},
		{"Succeeded", sourcegraph.BuildListOptions{Succeeded: true, Sort: "updated_at", Direction: "desc"}},
		{"Failed", sourcegraph.BuildListOptions{Failed: true, Sort: "updated_at", Direction: "desc"}},
	}

	pg, err := paginate(opt /* TODO */, 0)
	if err != nil {
		return err
	}

	buildsAndCommits, err := fetchCommitsForBuilds(ctx, builds.Builds)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "builds/builds.html", http.StatusOK, nil, &struct {
		BuildsAndCommits []buildAndCommit
		Opt              sourcegraph.BuildListOptions
		Tabs             []tab
		PageLinks        []pageLink

		tmpl.Common
	}{
		BuildsAndCommits: buildsAndCommits,
		Opt:              opt,
		Tabs:             tabs,
		PageLinks:        pg,
	})
}

type buildAndCommit struct {
	Build  *sourcegraph.Build
	Commit *vcs.Commit
}

func fetchCommitsForBuilds(ctx context.Context, builds []*sourcegraph.Build) ([]buildAndCommit, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	buildsAndCommits := make([]buildAndCommit, len(builds))
	par := parallel.NewRun(8)
	for i, b := range builds {
		i, b := i, b
		buildsAndCommits[i].Build = b
		par.Do(func() error {
			commit, err := cl.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{
				RepoSpec: sourcegraph.RepoSpec{URI: b.Repo},
				Rev:      b.CommitID,
				CommitID: b.CommitID,
			})
			if err != nil && grpc.Code(err) != codes.NotFound {
				// Tolerate not found (happens when the commit is
				// gc'd).
				return err
			}
			buildsAndCommits[i].Commit = commit
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}
	return buildsAndCommits, nil
}
