package coverage

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sync"

	"sourcegraph.com/sourcegraph/srclib/cvg"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

type Coverage struct {
	Repo string
	Cov  cvg.Coverage

	FileScoreClass  string
	RefScoreClass   string
	TokDensityClass string
}

type langCoverage struct {
	Lang    string
	RepoCov []*Coverage
}

func AddRoutes(r *router.Router) {
	r.Path("/.coverage").Methods("GET").Handler(internal.Handler(serveCoverage))
}

// serveCoverage serves a dashboard summarizing Sourcegraph's coverage of the top repositories in
// each language. By default, it shows the top 5 repositories in each language. You can also specify
// the optional `lang` URL parameter to show coverage for the top 100 repositories in a given
// language.
func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	if !auth.ActorFromContext(ctx).HasAdminAccess() {
		return fmt.Errorf("must be admin to access coverage dashboard")
	}

	langs, langRepos := make([]string, 0), make(map[string][]string)
	if specificRepo := r.URL.Query().Get("repo"); specificRepo != "" {
		langs = []string{"Repository coverage"}
		langRepos["Repository coverage"] = []string{specificRepo}
	} else if l := r.URL.Query().Get("lang"); l != "" {
		langs = []string{l}
		langRepos[l] = langRepos_[l]
	} else {
		langs = langs_
		for _, lang := range langs {
			repos := langRepos_[lang]
			if len(repos) > 5 {
				repos = repos[:5]
			}
			langRepos[lang] = repos
		}
	}

	var data struct {
		Langs []*langCoverage
		tmpl.Common
	}

	p := parallel.NewRun(10)
	var dlMu sync.Mutex

	for _, lang := range langs {
		lang := lang

		repos := langRepos[lang]
		langCov := &langCoverage{Lang: lang}
		data.Langs = append(data.Langs, langCov)
		for _, repo := range repos {
			repo := repo

			p.Do(func() error {
				cov, err := getCoverage(cl, ctx, repo)
				if err != nil {
					return err
				}
				{
					dlMu.Lock()
					defer dlMu.Unlock()
					langCov.RepoCov = append(langCov.RepoCov, cov)
				}
				return nil
			})
		}
	}
	if err := p.Wait(); err != nil {
		return err
	}

	return tmpl.Exec(r, w, "coverage/coverage.html", http.StatusOK, nil, &data)
}

func getCoverage(cl *sourcegraph.Client, ctx context.Context, repo string) (*Coverage, error) {
	repoSpec := sourcegraph.RepoSpec{URI: repo}
	commit, err := cl.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{RepoSpec: repoSpec})
	if err != nil {
		if handlerutil.IsRepoNoVCSDataError(err) {
			log15.Warn("getCoverage: no VCS data found, attempting to clone", "repo", repo)
			return &Coverage{Repo: repo, FileScoreClass: "fail", RefScoreClass: "fail", TokDensityClass: "fail"}, nil
		}
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
			var c cvg.Coverage
			err := json.Unmarshal([]byte(status.Description), &c)
			if err != nil {
				return nil, err
			}
			cov.Cov = c
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
