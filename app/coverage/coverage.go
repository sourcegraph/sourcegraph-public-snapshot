package coverage

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"

	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/srclib/cvg"
)

type Coverage struct {
	Repo string
	Cov  cvg.Coverage

	SuccessfullyBuilt bool
	FileScoreClass    string
	RefScoreClass     string
	TokDensityClass   string

	CommitsBehind int32
}

type langCoverage struct {
	Lang    string
	RepoCov []*Coverage
}

func AddRoutes(r *router.Router) {
	r.Path("/-/coverage").Methods("GET", "POST").Handler(internal.Handler(serveCoverage))
}

// serveCoverage serves a dashboard summarizing Sourcegraph's coverage of the top repositories in
// each language. By default, it shows the top 5 repositories in each language. You can also specify
// the optional `lang` URL parameter to show coverage for the top 100 repositories in a given
// language.
func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	rebuildMissing := false
	if r.Form.Get("process-empty") != "" {
		rebuildMissing = true
	}

	ctx, cl := handlerutil.Client(r)

	if !auth.ActorFromContext(ctx).HasAdminAccess() {
		return fmt.Errorf("must be admin to access coverage dashboard")
	}

	var langs []string
	langRepos := make(map[string][]string)
	if specificRepo := r.URL.Query().Get("repo"); specificRepo != "" {
		langs = []string{"Repository coverage"}
		langRepos["Repository coverage"] = []string{specificRepo}
	} else if l := r.URL.Query().Get("lang"); l != "" {
		langs = []string{l}
		langRepos[l] = langRepos_[l]
	} else {
		langs = append(langs, langs_...)
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
				cov, err := getCoverage(cl, ctx, repo, rebuildMissing)
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

	if r.Method == "POST" {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return nil
	}

	return tmpl.Exec(r, w, "coverage/coverage.html", http.StatusOK, nil, &data)
}

func getCoverage(cl *sourcegraph.Client, ctx context.Context, repo string, rebuildMissing bool) (*Coverage, error) {
	repoRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: repo},
	}
	// Query for srclib data with an empty path, which will force a lookback on the default branch
	// as far as necessary (upto a limit), until a built commit is found.
	dataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: repoRevSpec,
	})

	// Handle the potential error. If rebuildMissing, attempt to
	// generate the data, but return empty Coverage for now
	if err != nil {
		if rebuildMissing {
			if strings.Contains(err.Error(), "not found") && strings.Contains(err.Error(), "repo") {
				if err := createMirrorRepo(cl, ctx, repo); err != nil {
					return nil, err
				}
			} else if grpc.Code(err) == codes.NotFound {
				masterCommit, err := cl.Repos.GetCommit(ctx, &repoRevSpec)
				if err != nil {
					return nil, err
				}
				if _, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{
					Repo:     repoRevSpec.RepoSpec,
					CommitID: string(masterCommit.ID),
					Config:   sourcegraph.BuildConfig{Queue: true, Priority: 100},
				}); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		return &Coverage{Repo: repo, SuccessfullyBuilt: false}, nil
	}

	repoRevSpec.CommitID = dataVer.CommitID

	cstatus, err := cl.RepoStatuses.GetCombined(ctx, &repoRevSpec)
	if err != nil {
		return nil, err
	}

	var cov Coverage
	cov.Repo = repo
	cov.SuccessfullyBuilt = true
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
	cov.CommitsBehind = dataVer.CommitsBehind
	return &cov, nil
}

func createMirrorRepo(cl *sourcegraph.Client, ctx context.Context, repo string) error {
	// Resolve repo path, and create local mirror for remote repo if
	// needed.
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repo})
	if err != nil {
		return err
	}
	if remoteRepo := res.GetRemoteRepo(); remoteRepo != nil {
		if actualURI := githubutil.RepoURI(remoteRepo.Owner, remoteRepo.Name); actualURI != repo {
			// Repo path is invalid, possibly because repo has been renamed.
			return fmt.Errorf("repo %s redirects to %s; update dashboard with correct repo path", repo, actualURI)
		}

		// Automatically create a local mirror.
		log15.Info("Creating a local mirror of remote repo", "cloneURL", remoteRepo.HTTPCloneURL)
		_, err := cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_FromGitHubID{FromGitHubID: int32(remoteRepo.GitHubID)},
		})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("remote repo for resolution %+v was nil", res)
	}
	return nil
}
