package cli

import (
	"encoding/json"
	"fmt"
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

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/coverageutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/plan"
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

	cache = &coverageCache{
		DefsCache:              make(map[string]*defIndex),
		ResolvedRevCache:       make(map[string]string),
		SrclibDataVersionCache: make(map[string]string),
		FetchDefsMus:           make(map[string]*sync.Mutex),
	}
}

type coverageCmd struct {
	Repo       string `long:"repo" description:"repo URI"`
	Lang       string `long:"lang" description:"coverage language"`
	Refresh    bool   `long:"refresh" description:"refresh repo VCS data (or clone the repo if it doesn't exist); queue a new build"`
	Dry        bool   `long:"dry" description:"do a dry run (don't save coverage data)"`
	Progress   bool   `long:"progress" description:"show progress"`
	ReportRefs bool   `long:"refs" description:"report issues with references"`
	ReportDefs bool   `long:"defs" description:"report issues with definitions"`
}

// fileCoverage contains coverage data for a single file or repository
type fileCoverage struct {
	Path             string // the file path (optional)
	Idents           int    // # of identifiers in the file
	Refs             int    // # of refs in the file (i.e. annotations)
	Defs             int    // # of annotations (URLs) which resolve to real defs
	UnresolvedIdents []*coverageutil.Token
	UnresolvedRefs   []*coverageutil.Token
}

// repoCoverage contains the coverage data for a single repo
type repoCoverage struct {
	Repo          string
	Rev           string
	Language      string
	Files         []*fileCoverage
	Summary       *fileCoverage
	SrclibVersion string
	Day           string
	Duration      float64 // time to compute coverage (in seconds)
}

// defIndex contains all of the defs for a particular repo@commit
// indexed by key, and a mutex for concurrent access
type defIndex struct {
	Mu    sync.Mutex
	Index map[string]*sourcegraph.Def
}

// get is a threadsafe accessor function for a defIndex
func (idx *defIndex) get(key string) *sourcegraph.Def {
	idx.Mu.Lock()
	defer idx.Mu.Unlock()
	def, ok := idx.Index[key]
	if !ok {
		return nil
	}
	return def
}

// put is a threadsafe setter for defIndex
func (idx *defIndex) put(key string, def *sourcegraph.Def) {
	idx.Mu.Lock()
	defer idx.Mu.Unlock()
	idx.Index[key] = def
}

// coverageCache caches data fetched while computing
// coverage and includes mutexes for concurrent access
type coverageCache struct {
	DefsCacheMu            sync.Mutex
	DefsCache              map[string]*defIndex // key is repo@commit
	ResolvedRevMu          sync.Mutex
	ResolvedRevCache       map[string]string // repo => commit
	SrclibDataVersionMu    sync.Mutex
	SrclibDataVersionCache map[string]string // repo@rev => commit

	// FetchDefsMus allows at most one goroutine to fetch defs per repo@commit.
	// Map access is guarded by FetchDefsMu.
	FetchDefsMu  sync.Mutex
	FetchDefsMus map[string]*sync.Mutex
}

// cache is a global instance of coverageCache
var cache *coverageCache

// getFetchDefsMu acquires a lock to fetch defs for a repo@commit; it is threadsafe
func (c *coverageCache) getFetchDefsMu(key string) *sync.Mutex {
	c.FetchDefsMu.Lock()
	defer c.FetchDefsMu.Unlock()

	mu, ok := c.FetchDefsMus[key]
	if !ok {
		mu = &sync.Mutex{}
		c.FetchDefsMus[key] = mu
	}

	return mu
}

// getDefIndex is a threadsafe accessor function for cached def data
func (c *coverageCache) getDefIndex(key string) *defIndex {
	c.DefsCacheMu.Lock()
	defer c.DefsCacheMu.Unlock()
	idx, ok := c.DefsCache[key]
	if !ok {
		return nil
	}
	return idx
}

// putDefIndex is a threadsafe setter for cached def data
func (c *coverageCache) putDefIndex(key string, idx *defIndex) {
	c.DefsCacheMu.Lock()
	defer c.DefsCacheMu.Unlock()
	c.DefsCache[key] = idx
}

// getSrclibDataVersion returns (or fetches) the srclib data version
// for a particular repo@rev; it is threadsafe
func (c *coverageCache) getSrclibDataVersion(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec) string {
	c.SrclibDataVersionMu.Lock()
	defer c.SrclibDataVersionMu.Unlock()

	key := repoRevKey(repoRev)
	dataVer, ok := c.SrclibDataVersionCache[key]
	if !ok {
		sdv, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{RepoRev: *repoRev})
		if err != nil {
			log15.Debug("get srclib data version", "err", err)
			return ""
		}
		if sdv.CommitID == "" {
			log15.Debug("empty srclib data version", "err", err)
			return ""
		}
		dataVer = sdv.CommitID
	}

	c.SrclibDataVersionCache[key] = dataVer
	return dataVer
}

// getResolvedRev returns (or fetches) the absolute commit ID for the default branch
// for a particular repo; it is threadsafe
func (c *coverageCache) getResolvedRev(cl *sourcegraph.Client, ctx context.Context, repoPath string) (string, error) {
	c.ResolvedRevMu.Lock()
	defer c.ResolvedRevMu.Unlock()

	if commitID, ok := c.ResolvedRevCache[repoPath]; ok {
		return commitID, nil
	}

	repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoPath})
	if err != nil {
		return "", err
	}

	res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repoPath, Rev: repo.DefaultBranch})
	if err != nil {
		return "", err
	}

	c.ResolvedRevCache[repoPath] = res.CommitID
	return res.CommitID, nil
}

// fetchAndIndexDefs fetches (and indexes) all of the defs for a repo@rev, then caches the result.
// If the cache already contains data for repo@rev, it is returned immediately.
func (c *coverageCache) fetchAndIndexDefs(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec) *defIndex {
	// First resolve the rev to an absolute commit ID.
	repoRev.CommitID = c.getSrclibDataVersion(cl, ctx, repoRev)
	if repoRev.CommitID == "" {
		return nil
	}

	rr := repoRevKey(repoRev)

	fetchMu := c.getFetchDefsMu(rr)
	fetchMu.Lock()
	defer fetchMu.Unlock()

	if idx := c.getDefIndex(rr); idx != nil {
		return idx
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
		idx.put(defKey(&defSpec), def)
	}

	c.putDefIndex(rr, &idx)
	return &idx
}

// repoRevKey returns the cache key for a repo@commit
func repoRevKey(repoRev *sourcegraph.RepoRevSpec) string {
	return fmt.Sprintf("%s@%s", repoRev.Repo, repoRev.CommitID)
}

// defKey returns the cache key for a def
func defKey(def *sourcegraph.DefSpec) string {
	return fmt.Sprintf("%s/%s/-/%s", def.UnitType, def.Unit, def.Path)
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

	langCvg := make(map[string][]*repoCoverage)

	p := parallel.NewRun(30)
	var mu sync.Mutex

	for _, lang := range langs {

		lang := lang
		repos := langRepos[lang]
		for _, repo := range repos {
			repo := repo

			p.Do(func() error {
				cov, err := getCoverage(cl, cliContext, repo, lang, c.Dry, c.Progress, c.ReportRefs, c.ReportDefs)
				if err != nil {
					return fmt.Errorf("error getting coverage for %s: %s", repo, err)
				}
				{
					mu.Lock()
					defer mu.Unlock()
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

// rel is a *mux.Router for parsing vars from an annotation URL
var rel = router.New(nil)

// parseAnnotationURL extracts repoRev and def specs from an annotationURL
func parseAnnotationURL(annUrl string) (*sourcegraph.RepoRevSpec, *sourcegraph.DefSpec, error) {
	var match mux.RouteMatch
	if rel.Match(&http.Request{Method: "GET", URL: &url.URL{Path: fmt.Sprintf("/%s%s", "repos", annUrl)}}, &match) {
		repoRev := routevar.ToRepoRev(match.Vars)
		def := routevar.ToDefAtRev(match.Vars)
		return &sourcegraph.RepoRevSpec{
				Repo:     repoRev.Repo,
				CommitID: repoRev.Rev,
			}, &sourcegraph.DefSpec{
				Repo:     def.Repo,
				CommitID: def.Rev,
				UnitType: def.UnitType,
				Unit:     def.Unit,
				Path:     def.Path,
			}, nil
	} else {
		return nil, nil, fmt.Errorf("error parsing mux vars for annotation url %s", annUrl)
	}
}

// annToken stores an annotation (ref) and its associated token (ident)
type annToken struct {
	Annotation *sourcegraph.Annotation
	Token      *coverageutil.Token
}

// getFileCoverage computes the coverage data for a single file in a repository
func getFileCoverage(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, path, lang string, reportRefs, reportDefs bool) (*fileCoverage, error) {

	fileCvg := &fileCoverage{Path: path}

	var tokenizer coverageutil.Tokenizer
	if t := coverageutil.Lookup(lang, path); t != nil {
		tokenizer = *t
	} else {
		return nil, nil
	}

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
		return nil, err
	}

	anns, err := cl.Annotations.List(ctx, &sourcegraph.AnnotationsListOptions{
		Entry: entrySpec,
		Range: &sourcegraph.FileRange{StartByte: 0, EndByte: 0},
	})
	if err != nil {
		return nil, err
	}

	annsByStartByte := make(map[uint32]*sourcegraph.Annotation)
	for _, ann := range anns.Annotations {
		// require URL (i.e. don't count syntax highlighting annotations)
		if ann.URL != "" || len(ann.URLs) > 0 {
			annsByStartByte[ann.StartByte] = ann
		}
	}

	tokenizer.Init(entry.Contents)
	defer tokenizer.Done()

	refAnnotations := make([]*annToken, 0)
	for {
		tok := tokenizer.Next()
		if tok == nil {
			break
		}

		fileCvg.Idents += 1
		var hasRef bool
		if ann, ok := annsByStartByte[tok.Offset]; ok {
			if ann.EndByte == tok.Offset+uint32(len([]byte(tok.Text))) {
				// ref counts exact matches only
				fileCvg.Refs += 1
				refAnnotations = append(refAnnotations, &annToken{Annotation: ann, Token: tok})
				hasRef = true
			} else if reportRefs {
				log15.Warn("spans not match", "path", path, "at", tok.Offset, "ident", tok.Text)
			}
		} else if reportRefs {
			log15.Warn("no ref for", "path", path, "at", tok.Offset, "ident", tok.Text)
		}
		if !hasRef {
			fileCvg.UnresolvedIdents = append(fileCvg.UnresolvedIdents, tok)
		}
	}
	errors := tokenizer.Errors()
	if len(errors) > 0 {
		log15.Warn("parse errors", "path", path, "errors", errors)
	}

	for _, annToken := range refAnnotations {
		ann := annToken.Annotation
		tok := annToken.Token

		// Verify if the annotation (ref) resolves to a def.
		var u string
		if ann.URL != "" {
			u = ann.URL
		} else if ann.URLs[0] != "" {
			u = ann.URLs[0] // heuristic: just use first URL
		} else {
			continue
		}
		uStruct, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		if uStruct.IsAbs() {
			fileCvg.Defs += 1 // heuristic: consider absolute URLs to be valid defs
			continue
		}
		annRepoRev, annDefSpec, err := parseAnnotationURL(u)
		if err != nil {
			return nil, err
		}
		if annRepoRev.CommitID == "" {
			commitID, err := cache.getResolvedRev(cl, ctx, annRepoRev.Repo)
			if err != nil || commitID == "" {
				// The ref cannot be resolved to a def (e.g. the def repo doesn't exist);
				// this is a normal condition for the coverage script so swallow the error and continue.
				fileCvg.UnresolvedRefs = append(fileCvg.UnresolvedRefs, tok)
				continue
			}
			annRepoRev.CommitID = commitID
		}

		defIdx := cache.fetchAndIndexDefs(cl, ctx, annRepoRev)
		if defIdx == nil {
			fileCvg.UnresolvedRefs = append(fileCvg.UnresolvedRefs, tok)
			continue
		}
		if def := defIdx.get(defKey(annDefSpec)); def != nil {
			fileCvg.Defs += 1
		} else {
			fileCvg.UnresolvedRefs = append(fileCvg.UnresolvedRefs, tok)
			if reportDefs {
				log15.Warn("no def", "path", path, "at", ann.StartByte, "key", u)
			}
		}

	}

	// TEMPORARY: nillify unresolved idents / refs, to reduce storage impact.
	fileCvg.UnresolvedIdents = nil
	fileCvg.UnresolvedRefs = nil

	return fileCvg, nil
}

// getCoverage computes coverage data for the given repository
func getCoverage(cl *sourcegraph.Client, ctx context.Context, repoPath, lang string, dryRun, progress, reportRefs, reportDefs bool) (*repoCoverage, error) {
	if err := ensureRepoExists(cl, ctx, repoPath); err != nil {
		return nil, err
	}

	start := time.Now()

	repoCvg := repoCoverage{Repo: repoPath, Day: start.Format("01-02"), Language: lang}
	commitID, err := cache.getResolvedRev(cl, ctx, repoPath)
	if err != nil {
		return nil, err
	}

	repoRevSpec := sourcegraph.RepoRevSpec{Repo: repoPath, CommitID: commitID}
	dataVer := cache.getSrclibDataVersion(cl, ctx, &repoRevSpec)
	if dataVer != "" {
		repoRevSpec.CommitID = dataVer
		repoCvg.Rev = dataVer
		tree, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: repoRevSpec})
		if err != nil {
			return nil, err
		}

		p := parallel.NewRun(10)
		var repoCvgMu sync.Mutex
		for _, path := range tree.Files {
			path := path

			p.Do(func() error {
				if progress {
					log15.Info("processing", path, lang)
				}
				fileCvg, err := getFileCoverage(cl, ctx, &repoRevSpec, path, lang, reportRefs, reportDefs)
				if err != nil {
					return err
				}
				if fileCvg != nil {
					// fileCvg may be nil for files which are ignored / not indexed
					repoCvgMu.Lock()
					repoCvg.Files = append(repoCvg.Files, fileCvg)
					repoCvgMu.Unlock()
				}

				return nil
			})
		}

		if err := p.Wait(); err != nil {
			if errs, ok := err.(parallel.Errors); ok {
				var errMsgs []string
				for _, e := range errs {
					errMsgs = append(errMsgs, e.Error())
				}
				log15.Error("error computing coverage", "repo", repoPath, "err", fmt.Sprintf("\n%s", strings.Join(errMsgs, "\n")))
			}
		}
	} else {
		log15.Warn("missing srclib data version", "repo", repoPath, "rev", commitID)
	}

	repoCvg.Summary = &fileCoverage{}
	for _, cv := range repoCvg.Files {
		repoCvg.Summary.Idents += cv.Idents
		repoCvg.Summary.Refs += cv.Refs
		repoCvg.Summary.Defs += cv.Defs
	}
	log15.Info("coverage summary",
		"lang", lang,
		"repo", repoPath,
		"idents", repoCvg.Summary.Idents,
		"refs", repoCvg.Summary.Refs,
		"defs", repoCvg.Summary.Defs,
		"refs%", percent(repoCvg.Summary.Refs, repoCvg.Summary.Idents),
		"defs%", percent(repoCvg.Summary.Defs, repoCvg.Summary.Refs))

	if planVer, err := plan.SrclibVersion(lang); err != nil {
		log15.Warn("missing plan srclib version", "err", err, "lang", lang)
	} else {
		repoCvg.SrclibVersion = planVer
	}

	repoCvg.Duration = time.Since(start).Seconds()
	log15.Info("coverage duration", "seconds", repoCvg.Duration)

	if dryRun {
		return &repoCvg, nil
	}

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

// percent computes percentage safely
func percent(a, b int) int {
	if b == 0 {
		return 0
	}
	return a * 100.0 / b
}
