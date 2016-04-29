package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
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

func init() {
	_, err := cli.CLI.AddCommand("coverage",
		"get srclib coverage stats",
		"Retrieve the coverage stats for repos/commits indexed by Sourcegraph",
		&coverageCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type coverageCmd struct {
	Repo    string `long:"repo" description:"repo URI"`
	Lang    string `long:"lang" description:"coverage language"`
	Refresh bool   `long:"refresh" description:"refresh the coverage information or compute it if it doesn't exist yet"`
}

func (c *coverageCmd) Execute(args []string) error {
	cl := cliClient
	var langs []string
	langRepos := make(map[string][]string)
	if specificRepo := c.Repo; specificRepo != "" {
		if c.Lang == "" {
			return fmt.Errorf("must specify language")
		}
		langs = []string{c.Lang}
		langRepos[c.Lang] = []string{specificRepo}
	} else if l := c.Lang; l != "" {
		langs = []string{l}
		langRepos[l] = langRepos_[l]
	} else {
		// select top 5 repos for each lang
		langs = append(langs, langs_...)
		for _, lang := range langs {
			repos := langRepos_[lang]
			if len(repos) > 5 {
				repos = repos[:5]
			}
			langRepos[lang] = repos
		}
	}

	// If c.Refresh, then just call `src repo sync` for every repo
	if c.Refresh {
		var allRepos []string
		for _, repos := range langRepos {
			allRepos = append(allRepos, repos...)
		}

		syncCmd := &repoSyncCmd{Force: true}
		syncCmd.Args.URIs = allRepos
		return syncCmd.Execute(nil)
	}

	// Create coverage dashboard
	var data struct {
		Langs    []*langCoverage
		Endpoint string
	}
	data.Endpoint = endpoint.URL

	p := parallel.NewRun(30)
	var dlMu sync.Mutex

	for _, lang := range langs {
		lang := lang

		repos := langRepos[lang]
		langCov := &langCoverage{Lang: lang}
		data.Langs = append(data.Langs, langCov)
		for _, repo := range repos {
			repo := repo

			p.Do(func() error {
				cov, err := getCoverage(cl, cliContext, repo, lang)
				if err != nil {
					return fmt.Errorf("error getting coverage for %s: %s", repo, err)
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
		if errs, ok := err.(parallel.Errors); ok {
			var errMsgs []string
			for _, e := range errs {
				errMsgs = append(errMsgs, e.Error())
			}
			err = fmt.Errorf("\n%s", strings.Join(errMsgs, "\n"))
		}
		return fmt.Errorf("coverage errors: %s", err)
	}

	if err := coverageTemplate.Execute(os.Stdout, &data); err != nil {
		return err
	}

	return nil
}

func getCoverage(cl *sourcegraph.Client, ctx context.Context, repo string, lang string) (*Coverage, error) {
	repoRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: repo},
	}

	if err := ensureRepoExists(cl, ctx, repo); err != nil {
		return nil, err
	}

	var cov Coverage
	cov.Repo = repo
	cov.SuccessfullyBuilt = true
	langCov, dataVer, err := handlerutil.GetCoverage(cl, ctx, repoRevSpec.URI)
	if err == handlerutil.ErrCovNotExist {
		return &Coverage{Repo: repo, SuccessfullyBuilt: false}, nil
	} else if err != nil {
		return nil, err
	}
	if dataVer != nil {
		cov.CommitsBehind = dataVer.CommitsBehind
	}
	if lc := langCov[lang]; lc != nil {
		cov.Cov = *lc
	}
	cov.FileScoreClass, cov.RefScoreClass, cov.TokDensityClass = "fail", "fail", "fail"
	if cov.Cov.FileScorePass() {
		cov.FileScoreClass = "success"
	}
	if cov.Cov.RefScorePass() {
		cov.RefScoreClass = "success"
	}
	if cov.Cov.TokDensityPass() {
		cov.TokDensityClass = "success"
	}
	return &cov, nil
}

func ensureRepoExists(cl *sourcegraph.Client, ctx context.Context, repo string) error {
	// Resolve repo path, and create local mirror for remote repo if
	// needed.
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repo})
	if err != nil && grpc.Code(err) != codes.NotFound {
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
			Op: &sourcegraph.ReposCreateOp_FromGitHubID{FromGitHubID: remoteRepo.GitHubID},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
