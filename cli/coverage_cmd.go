package cli

import (
	"encoding/json"
	"fmt"
	"go/scanner"
	"go/token"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/plan"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
)

func init() {
	_, err := cli.CLI.AddCommand("coverage",
		"generate srclib coverage stats",
		"Compute coverage stats for repos/commits indexed by Sourcegraph; or sync repos to prepare coverage script",
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

// fileCoverage contains the coverage data for a single file
type fileCoverage struct {
	Path     string // the file path
	Language string // the language name
	Idents   int    // # of identifiers in the file
	Refs     int    // # of refs in the file (has annotations)
	Defs     int    // # of annotations which resolve the real defs
}

// srclibVersion contains the current version of a language toolchain
// at the time of running the script
type srclibVersion struct {
	Language string
	Version  string
}

// repoCoverage contains the coverage data for a single repo
type repoCoverage struct {
	Repo           string
	Files          []*fileCoverage
	Summary        []*fileCoverage // summation over Files, by language
	SrclibVersions []*srclibVersion
	Day            string
	Duration       float64 // time to compute coverage, in seconds
}

// defIndex contains all of the defs for a particular repo@commit
// indexed by key as well as a mutex for concurrent access
type defIndex struct {
	Mu    sync.Mutex
	Index map[string]*sourcegraph.Def
}

// coverageCache caches all of the data fetched while computing
// coverage and includes mutexes for concurrent access
type coverageCache struct {
	DefsCacheMu            sync.Mutex
	DefsCache              map[string]*defIndex // key is repo@commit
	SrclibDataVersionMu    sync.Mutex
	SrclibDataVersionCache map[string]string // repo@rev => commit
}

// newCoverageCache creates a new coverageCache
func newCoverageCache() *coverageCache {
	return &coverageCache{
		DefsCache:              make(map[string]*defIndex),
		SrclibDataVersionCache: make(map[string]string),
	}
}

// repoRevKey returns the cache key for a repo@rev
func repoRevKey(repoRev *sourcegraph.RepoRevSpec) string {
	return fmt.Sprintf("%s@%s", repoRev.RepoSpec.URI, repoRev.Rev)
}

// repoCommitKey returns the cache key for a repo@commit
func repoCommitKey(repoRev *sourcegraph.RepoRevSpec) string {
	return fmt.Sprintf("%s@%s", repoRev.RepoSpec.URI, repoRev.CommitID)
}

// defKey returns the cache key for a def
func defKey(def *sourcegraph.DefSpec) string {
	return fmt.Sprintf("%s/%s/-/%s", def.UnitType, def.Unit, def.Path)
}

// rel is a *mux.Router for parsing vars from an annotation URL
var rel = router.New(nil)

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

	langCvg := make(map[string][]*repoCoverage)

	cvgCache := newCoverageCache()
	p := parallel.NewRun(30)
	var dlMu sync.Mutex

	for _, lang := range langs {
		repos := langRepos[lang]
		for _, repo := range repos {
			repo := repo

			p.Do(func() error {
				cov, err := getCoverage(cl, cliContext, repo, lang, cvgCache)
				if err != nil {
					return fmt.Errorf("error getting coverage for %s: %s", repo, err)
				}
				{
					dlMu.Lock()
					defer dlMu.Unlock()
					if _, ok := langCvg[lang]; !ok {
						langCvg[lang] = make([]*repoCoverage, 0)
					}
					langCvg[lang] = append(langCvg[lang], cov)
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

	return nil
}

// parseAnnotationURL extracts repoRev and def specs from an annotationURL
func parseAnnotationURL(annUrl string) (*sourcegraph.RepoRevSpec, *sourcegraph.DefSpec, error) {
	var match mux.RouteMatch
	if rel.Match(&http.Request{Method: "GET", URL: &url.URL{Path: fmt.Sprintf("/%s%s", "repos", annUrl)}}, &match) {
		repoRev, err := sourcegraph.UnmarshalRepoRevSpec(match.Vars)
		if err != nil {
			return nil, nil, err
		}

		defSpec, err := sourcegraph.UnmarshalDefSpec(match.Vars)
		if err != nil {
			return nil, nil, err
		}

		return &repoRev, &defSpec, nil
	} else {
		return nil, nil, fmt.Errorf("error parsing mux vars for annotation url %s", annUrl)
	}
}

// fetchAndCacheDefs fetches (and indexes) all of the defs for a repo@rev, then caches the result,
// if and only if the cache does not already contain data for repo@rev.
func fetchAndCacheDefs(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, c *coverageCache) {
	// To fetch all defs we must resolve the rev to an absolute commit ID.
	// Keep this value cached to avoid unnecessary round trips to the same repo.
	c.SrclibDataVersionMu.Lock()
	if _, ok := c.SrclibDataVersionCache[repoRevKey(repoRev)]; !ok {
		dataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{RepoRev: *repoRev})
		if err != nil {
			log15.Debug("get srclib data version", "err", err)
			c.SrclibDataVersionMu.Unlock()
			return
		}
		if dataVer.CommitID == "" {
			log15.Debug("empty srclib data version", "err", err)
			c.SrclibDataVersionMu.Unlock()
			return
		}

		c.SrclibDataVersionCache[repoRevKey(repoRev)] = dataVer.CommitID
	}

	commitID, _ := c.SrclibDataVersionCache[repoRevKey(repoRev)]
	c.SrclibDataVersionMu.Unlock()

	repoRev.CommitID = commitID
	rr := repoCommitKey(repoRev) // repo@commit

	c.DefsCacheMu.Lock()
	defer c.DefsCacheMu.Unlock()

	if _, ok := c.DefsCache[rr]; ok {
		return
	}

	opt := sourcegraph.DefListOptions{
		IncludeTest: true,
		RepoRevs:    []string{rr},
	}
	opt.PerPage = 100000000 // TODO(rothfels): srclib def store doesn't properly handle pagination
	opt.Page = 1

	defs := make([]*sourcegraph.Def, 0)
	for {
		dl, err := cl.Defs.List(ctx, &opt)
		if err != nil {
			log15.Error("fetch defs", "err", err, "repoRev", rr)
			break
		}
		if len(dl.Defs) == 0 {
			break
		}
		defs = append(defs, dl.Defs...)
		opt.Page += 1
	}

	idx := defIndex{Index: make(map[string]*sourcegraph.Def)}
	for _, def := range defs {
		defSpec := def.DefSpec()
		idx.Index[defKey(&defSpec)] = def
	}

	c.DefsCache[rr] = &idx
}

// runCoverage computes the coverage data for a single file within a repository
func runCoverage(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, path, lang string, c *coverageCache, repoCvg *repoCoverage, p *parallel.Run, mu *sync.Mutex) {
	p.Do(func() error {
		// TODO(rothfels): add other language support.
		switch lang {
		case "Go":
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
		default:
			return nil
		}

		fileCvg := &fileCoverage{Path: path, Language: "Go"}
		mu.Lock()
		repoCvg.Files = append(repoCvg.Files, fileCvg)
		mu.Unlock()

		entrySpec := sourcegraph.TreeEntrySpec{
			RepoRev: *repoRev,
			Path:    path,
		}
		treeGetOp := sourcegraph.RepoTreeGetOptions{}
		treeGetOp.EntireFile = true
		entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
			Entry: entrySpec,
			Opt:   &treeGetOp,
		})
		if err != nil {
			return err
		}

		anns, err := cl.Annotations.List(ctx, &sourcegraph.AnnotationsListOptions{
			Entry: entrySpec,
			Range: &sourcegraph.FileRange{StartByte: 0, EndByte: 0},
		})
		if err != nil {
			return err
		}

		annsByStartByte := make(map[uint32]*sourcegraph.Annotation)
		for _, ann := range anns.Annotations {
			// require URL (i.e. don't count syntax highlighting annotations)
			if ann.URL != "" || len(ann.URLs) > 0 {
				annsByStartByte[ann.StartByte+1] = ann // off by one?
			}
		}

		scanner := &scanner.Scanner{}
		fset := token.NewFileSet()
		file := fset.AddFile(``, fset.Base(), len(entry.Contents))
		scanner.Init(file, entry.Contents, nil /* no error handler */, 0)

		refAnnotations := make([]*sourcegraph.Annotation, 0)
		for {
			pos, tok, _ := scanner.Scan()
			if tok == token.EOF {
				break
			}
			if tok != token.IDENT {
				continue
			}

			fileCvg.Idents += 1
			if ann, ok := annsByStartByte[uint32(pos)]; ok {
				fileCvg.Refs += 1
				refAnnotations = append(refAnnotations, ann)
			}
		}

		for _, ann := range refAnnotations {
			// Verify if the annotation (ref) resolves to a def.
			var u string
			if ann.URL != "" {
				u = ann.URL
			} else if ann.URLs[0] != "" {
				u = ann.URLs[0] // heuristic: just use first URL
			} else {
				continue
			}

			annRepoRev, annDefSpec, err := parseAnnotationURL(u)
			if err != nil {
				return nil
			}

			fetchAndCacheDefs(cl, ctx, annRepoRev, c)

			c.DefsCacheMu.Lock()
			defIdx, ok := c.DefsCache[repoCommitKey(annRepoRev)]
			if !ok {
				c.DefsCacheMu.Unlock()
				continue
			}
			c.DefsCacheMu.Unlock()

			defIdx.Mu.Lock()
			if _, ok := defIdx.Index[defKey(annDefSpec)]; ok {
				fileCvg.Defs += 1
			}
			defIdx.Mu.Unlock()
		}

		log15.Info("computed coverage", "path", path, "idents", fileCvg.Idents, "refs", fileCvg.Refs, "defs", fileCvg.Defs)
		return nil
	})
}

// getCoverage computes coverage data for the given repository
func getCoverage(cl *sourcegraph.Client, ctx context.Context, repoURI, lang string, c *coverageCache) (*repoCoverage, error) {
	if err := ensureRepoExists(cl, ctx, repoURI); err != nil {
		return nil, err
	}

	start := time.Now()

	repoSpec := sourcegraph.RepoSpec{URI: repoURI}
	repo, err := cl.Repos.Get(ctx, &repoSpec)
	if err != nil {
		return nil, err
	}

	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, Rev: repo.DefaultBranch}
	dataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{RepoRev: repoRevSpec})
	if err != nil {
		return nil, err
	}
	if dataVer.CommitID == "" {
		return nil, fmt.Errorf("missing srclib data version for %s@%s", repoURI, repo.DefaultBranch)
	}

	c.SrclibDataVersionMu.Lock()
	c.SrclibDataVersionCache[repoRevKey(&repoRevSpec)] = dataVer.CommitID
	c.SrclibDataVersionMu.Unlock()

	repoRevSpec.CommitID = dataVer.CommitID

	tree, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: repoRevSpec})
	if err != nil {
		return nil, err
	}

	p := parallel.NewRun(10)
	var repoCvgMu sync.Mutex
	repoCvg := repoCoverage{Repo: repoURI, Day: start.Format("01-02")}
	for _, path := range tree.Files {
		runCoverage(cl, ctx, &repoRevSpec, path, lang, c, &repoCvg, p, &repoCvgMu)
	}

	if err := p.Wait(); err != nil {
		if errs, ok := err.(parallel.Errors); ok {
			var errMsgs []string
			for _, e := range errs {
				errMsgs = append(errMsgs, e.Error())
			}
			err = fmt.Errorf("\n%s", strings.Join(errMsgs, "\n"))
		}
		return nil, fmt.Errorf("coverage errors: %s", err)
	}

	summaryByLang := make(map[string]*fileCoverage)
	for _, cv := range repoCvg.Files {
		if _, ok := summaryByLang[cv.Language]; !ok {
			summaryByLang[cv.Language] = &fileCoverage{Language: cv.Language}
		}
		summaryByLang[cv.Language].Idents += cv.Idents
		summaryByLang[cv.Language].Refs += cv.Refs
		summaryByLang[cv.Language].Defs += cv.Defs
	}

	summary := make([]*fileCoverage, 0)
	srclibVersions := make([]*srclibVersion, 0)
	for lang, cv := range summaryByLang {
		summary = append(summary, cv)
		v, err := plan.SrclibVersion(lang)
		if err != nil {
			log15.Warn("missing srclib version", "err", err, "lang", lang)
		}
		srclibVersions = append(srclibVersions, &srclibVersion{Language: lang, Version: v})

		log15.Info("coverage summary", "lang", lang, "repo", repoURI, "idents", cv.Idents, "refs", cv.Refs, "defs", cv.Defs)
	}

	repoCvg.Summary = summary
	repoCvg.SrclibVersions = srclibVersions
	repoCvg.Duration = time.Since(start).Seconds()
	log15.Info("coverage duration", "seconds", repoCvg.Duration)

	covJSON, err := json.Marshal([]repoCoverage{repoCvg})
	if err != nil {
		return nil, err
	}

	var statusUpdate sourcegraph.RepoStatusesCreateOp
	repoRevSpec.CommitID = "" // overwrite so db entry is saved with branch rev
	statusUpdate.Repo = repoRevSpec
	statusUpdate.Status = sourcegraph.RepoStatus{
		Context:     "coverage",
		Description: string(covJSON),
	}

	if _, err = cl.RepoStatuses.Create(ctx, &statusUpdate); err != nil {
		return nil, err
	}

	return &repoCvg, nil
}

func ensureRepoExists(cl *sourcegraph.Client, ctx context.Context, repo string) error {
	// Resolve repo path, and create local mirror for remote repo if needed.
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
