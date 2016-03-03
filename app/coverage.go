package app

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/srclib/cli"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

var langs = []string{"Go", "Java"}
var langRepos = map[string][]string{
	"Go": []string{
		"github.com/gorilla/mux",
	},
	"Java": []string{},
}

type Coverage struct {
	Repo string
	Cov  cli.Coverage

	FileScoreClass  string
	RefScoreClass   string
	TokDensityClass string
}

type langCoverage struct {
	Lang    string
	RepoCov []*Coverage
}

func getCoverage(cl *sourcegraph.Client, ctx context.Context, repo string) (*Coverage, error) {
	repoSpec := sourcegraph.RepoSpec{URI: repo}
	commit, err := cl.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{RepoSpec: repoSpec})
	if err != nil {
		return nil, err
	}

	cstatus, err := cl.RepoStatuses.GetCombined(ctx, &sourcegraph.RepoRevSpec{RepoSpec: repoSpec, CommitID: string(commit.ID)})
	if err != nil {
		return nil, err
	}

	var cov Coverage
	cov.Repo = repo
	for _, status := range cstatus.Statuses {
		if status.Context == "coverage" {
			var cvg cli.Coverage
			err := json.Unmarshal([]byte(status.Description), &cvg)
			if err != nil {
				return nil, err
			}
			cov.Cov = cvg
			break
		}
	}
	cov.FileScoreClass, cov.RefScoreClass, cov.TokDensityClass = "fail", "fail", "fail"
	if cov.Cov.FileScore > 0.8 {
		cov.FileScoreClass = "success"
	}
	if cov.Cov.RefScore > 0.95 {
		cov.RefScoreClass = "success"
	}
	if cov.Cov.TokDensity > 1.0 {
		cov.TokDensityClass = "success"
	}
	return &cov, nil
}

func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	var data struct {
		Langs []*langCoverage
		tmpl.Common
	}

	for _, lang := range langs {
		repos := langRepos[lang]

		langCov := &langCoverage{Lang: lang}
		for _, repo := range repos {
			cov, err := getCoverage(cl, ctx, repo)
			if err != nil {
				return err
			}
			langCov.RepoCov = append(langCov.RepoCov, cov)

		}
		data.Langs = append(data.Langs, langCov)
	}

	return tmpl.Exec(r, w, "coverage/coverage.html", http.StatusOK, nil, &data)
}
