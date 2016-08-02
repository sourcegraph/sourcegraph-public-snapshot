package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/neelance/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage/tokenizer"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/plan"
)

func init() {
	_, err := cli.CLI.AddCommand("srclib-coverage",
		"generate srclib coverage stats",
		"Compute coverage stats for repos/commits indexed by Sourcegraph; or sync repos to prepare coverage script",
		&srclibCoverageCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	cache = &coverageCache{
		DefsCache:              make(map[sourcegraph.RepoRevSpec]*defIndex),
		ResolvedRevCache:       make(map[routevar.RepoRev]sourcegraph.RepoRevSpec),
		SrclibDataVersionCache: make(map[sourcegraph.RepoRevSpec]string),
		FetchDefsMus:           make(map[sourcegraph.RepoRevSpec]*sync.Mutex),
	}
}

type srclibCoverageCmd struct {
	Repo        string `long:"repo" description:"repo URI"`
	Lang        string `long:"lang" description:"coverage language"`
	Limit       int    `long:"limit" description:"max number of repos to run coverage for"`
	Refresh     bool   `long:"refresh" description:"refresh repo VCS data (or clone the repo if it doesn't exist); queue a new build"`
	Dry         bool   `long:"dry" description:"do a dry run (don't save coverage data)"`
	Progress    bool   `long:"progress" description:"show progress"`
	ReportRefs  bool   `long:"refs" description:"report issues with references"`
	ReportDefs  bool   `long:"defs" description:"report issues with definitions"`
	ReportEmpty bool   `long:"empty" description:"report empty files"`
}

// srclibFileCoverage contains coverage data for a single file or repository
type srclibFileCoverage struct {
	Path   string // the file path (optional)
	Idents int    // # of identifiers in the file
	Refs   int    // # of refs in the file (i.e. annotations)
	Defs   int    // # of annotations (URLs) which resolve to real defs
}

// srclibRepoCoverage contains the coverage data for a single repo
type srclibRepoCoverage struct {
	Repo          string
	Rev           string
	Language      string
	Files         []*srclibFileCoverage
	Summary       *srclibFileCoverage
	SrclibVersion string
	Day           string
	Duration      float64 // time to compute coverage (in seconds)
}

// defIndex contains all of the defs for a particular repo@commit
// indexed by key, and a mutex for concurrent access
type defIndex struct {
	Mu    sync.Mutex
	Index map[sourcegraph.DefSpec]*sourcegraph.Def
}

// get is a threadsafe accessor function for a defIndex
func (idx *defIndex) get(key sourcegraph.DefSpec) *sourcegraph.Def {
	idx.Mu.Lock()
	defer idx.Mu.Unlock()
	def, ok := idx.Index[key]
	if !ok {
		return nil
	}
	return def
}

// put is a threadsafe setter for defIndex
func (idx *defIndex) put(key sourcegraph.DefSpec, def *sourcegraph.Def) {
	idx.Mu.Lock()
	defer idx.Mu.Unlock()
	idx.Index[key] = def
}

// coverageCache caches data fetched while computing
// coverage and includes mutexes for concurrent access
type coverageCache struct {
	DefsCacheMu            sync.Mutex
	DefsCache              map[sourcegraph.RepoRevSpec]*defIndex // key is repo@commit
	ResolvedRevMu          sync.Mutex
	ResolvedRevCache       map[routevar.RepoRev]sourcegraph.RepoRevSpec // (repo path, rev) => (repo ID, commit)
	SrclibDataVersionMu    sync.Mutex
	SrclibDataVersionCache map[sourcegraph.RepoRevSpec]string // repo@rev => commit

	// FetchDefsMus allows at most one goroutine to fetch defs per repo@commit.
	// Map access is guarded by FetchDefsMu.
	FetchDefsMu  sync.Mutex
	FetchDefsMus map[sourcegraph.RepoRevSpec]*sync.Mutex
}

// cache is a global instance of coverageCache
var cache *coverageCache

// getFetchDefsMu acquires a lock to fetch defs for a repo@commit; it is threadsafe
func (c *coverageCache) getFetchDefsMu(key sourcegraph.RepoRevSpec) *sync.Mutex {
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
func (c *coverageCache) getDefIndex(key sourcegraph.RepoRevSpec) *defIndex {
	c.DefsCacheMu.Lock()
	defer c.DefsCacheMu.Unlock()
	idx, ok := c.DefsCache[key]
	if !ok {
		return nil
	}
	return idx
}

// putDefIndex is a threadsafe setter for cached def data
func (c *coverageCache) putDefIndex(key sourcegraph.RepoRevSpec, idx *defIndex) {
	c.DefsCacheMu.Lock()
	defer c.DefsCacheMu.Unlock()
	c.DefsCache[key] = idx
}

// getSrclibDataVersion returns (or fetches) the srclib data version
// for a particular repo@rev; it is threadsafe
func (c *coverageCache) getSrclibDataVersion(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec) string {
	c.SrclibDataVersionMu.Lock()
	defer c.SrclibDataVersionMu.Unlock()

	dataVer, ok := c.SrclibDataVersionCache[*repoRev]
	if !ok {
		sdv, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{RepoRev: *repoRev})
		if err != nil {
			log15.Debug("get srclib data version", "err", err)
		} else if sdv.CommitID == "" {
			log15.Debug("empty srclib data version", "err", err)
		} else {
			dataVer = sdv.CommitID
		}
	} else {
		return dataVer
	}

	c.SrclibDataVersionCache[*repoRev] = dataVer
	return dataVer
}

// getResolvedRev returns (or fetches) the absolute commit ID for the default branch
// for a particular repo; it is threadsafe
func (c *coverageCache) getResolvedRev(cl *sourcegraph.Client, ctx context.Context, repoRev routevar.RepoRev) (sourcegraph.RepoRevSpec, error) {
	c.ResolvedRevMu.Lock()
	defer c.ResolvedRevMu.Unlock()

	key := repoRev
	if v, ok := c.ResolvedRevCache[key]; ok {
		return v, nil
	}

	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoRev.Repo})
	if err != nil {
		return sourcegraph.RepoRevSpec{}, err
	}

	if repoRev.Rev == "" {
		// Assume default branch is master to prevent call to Repos.Get.
		// This may break for some repos (in which case we may want to hardcode mappings
		// for exception cases).
		repoRev.Rev = "master"
	}

	resRev, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: res.Repo, Rev: repoRev.Rev})
	if err != nil {
		return sourcegraph.RepoRevSpec{}, err
	}

	v := sourcegraph.RepoRevSpec{Repo: res.Repo, CommitID: resRev.CommitID}
	c.ResolvedRevCache[key] = v
	return v, nil
}

// fetchAndIndexDefs fetches (and indexes) all of the defs for a repo@rev, then caches the result.
// If the cache already contains data for repo@rev, it is returned immediately.
func (c *coverageCache) fetchAndIndexDefs(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, repoURI string) *defIndex {
	// First resolve the rev to an absolute commit ID.
	repoRev.CommitID = c.getSrclibDataVersion(cl, ctx, repoRev)
	if repoRev.CommitID == "" {
		return nil
	}

	fetchMu := c.getFetchDefsMu(*repoRev)
	fetchMu.Lock()
	defer fetchMu.Unlock()

	if idx := c.getDefIndex(*repoRev); idx != nil {
		return idx
	}

	opt := sourcegraph.DefListOptions{
		IncludeTest: true,
		RepoRevs:    []string{fmt.Sprintf("%s@%s", repoURI, repoRev.CommitID)},
	}
	opt.PerPage = 100000000 // TODO(rothfels): srclib def store doesn't properly handle pagination
	opt.Page = 1

	defs := make([]*sourcegraph.Def, 0)
	for {
		dl, err := cl.Defs.List(ctx, &opt)
		if err != nil {
			log15.Error("fetch defs", "err", err, "repoRev", *repoRev)
			break
		}
		if len(dl.Defs) == 0 {
			break
		}
		defs = append(defs, dl.Defs...)
		opt.Page += 1
	}

	idx := defIndex{Index: make(map[sourcegraph.DefSpec]*sourcegraph.Def)}
	for _, def := range defs {
		defSpec := def.DefSpec(repoRev.Repo)
		idx.put(defSpec, def)
	}

	c.putDefIndex(*repoRev, &idx)
	return &idx
}

func (c *srclibCoverageCmd) Execute(args []string) error {
	cl := cliClient
	if c.Lang == "" {
		return fmt.Errorf("must specify language")
	}

	var repos []string
	if specificRepo := c.Repo; specificRepo != "" {
		repos = []string{specificRepo}
	} else {
		repos = langRepos[c.Lang]
	}

	if c.Limit > 0 && len(repos) > c.Limit {
		repos = repos[:c.Limit]
	}

	// If c.Refresh, then just call `src repo sync` for every repo
	if c.Refresh {
		slack.PostMessage(slack.PostOpts{
			Msg:        fmt.Sprintf("Running coverage --refresh --lang=%s", c.Lang),
			IconEmoji:  ":sourcegraph:",
			Channel:    "global-graph",
			WebhookURL: os.Getenv("SG_SLACK_GRAPH_WEBHOOK_URL"),
		})

		syncCmd := &repoSyncCmd{
			Force:         true,
			buildPriority: 0, // prioritize over background updates
		}
		syncCmd.Args.URIs = repos
		err := syncCmd.Execute(nil)
		if err != nil {
			log15.Error("repo sync", "err", err)
		}
		return err
	}

	start := time.Now()
	slack.PostMessage(slack.PostOpts{
		Msg:        fmt.Sprintf("Running coverage --lang=%s", c.Lang),
		IconEmoji:  ":chart_with_upwards_trend:",
		Channel:    "global-graph",
		WebhookURL: os.Getenv("SG_SLACK_GRAPH_WEBHOOK_URL"),
	})

	p := parallel.NewRun(30)
	for _, repo := range repos {
		repo := repo
		p.Acquire()
		go func() {
			defer p.Release()
			_, err := getCoverage(cl, cliContext, repo, c.Lang, c.Dry, c.Progress, c.ReportRefs, c.ReportDefs, c.ReportEmpty)
			if err != nil {
				p.Error(fmt.Errorf("error getting coverage for %s: %s", repo, err))
				return
			}
		}()
	}
	err := p.Wait()

	slack.PostMessage(slack.PostOpts{
		Msg:        fmt.Sprintf("Completed coverage --lang=%s (duration: %f mins)", c.Lang, time.Since(start).Minutes()),
		IconEmoji:  ":checkered-flag:",
		Channel:    "global-graph",
		WebhookURL: os.Getenv("SG_SLACK_GRAPH_WEBHOOK_URL"),
	})

	if err != nil {
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
func parseAnnotationURL(annUrl string) (routevar.DefAtRev, error) {
	var match mux.RouteMatch
	if rel.Match(&http.Request{Method: "GET", URL: &url.URL{Path: fmt.Sprintf("/%s%s", "repos", annUrl)}}, &match) {
		return routevar.ToDefAtRev(match.Vars), nil
	}
	return routevar.DefAtRev{}, fmt.Errorf("error parsing mux vars for annotation url %s", annUrl)
}

// annToken stores an annotation (ref) and its associated token (ident)
type annToken struct {
	Annotation *sourcegraph.Annotation
	Token      *tokenizer.Token
}

// getFileCoverage computes the coverage data for a single file in a repository
func getFileCoverage(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, repoPath, path, lang string, reportRefs, reportDefs, reportEmpty bool) (*srclibFileCoverage, error) {
	fileCvg := &srclibFileCoverage{Path: path}

	var tt tokenizer.Tokenizer
	if t := tokenizer.Lookup(lang, path); t != nil {
		tt = *t
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

	tt.Init(entry.Contents)
	defer tt.Done()

	refAnnotations := make([]*annToken, 0)
	for {
		tok := tt.Next()
		if tok == nil {
			break
		}

		fileCvg.Idents += 1
		if ann, ok := annsByStartByte[tok.Offset]; ok {
			if ann.EndByte == tok.Offset+uint32(len([]byte(tok.Text))) {
				// ref counts exact matches only
				fileCvg.Refs += 1
				refAnnotations = append(refAnnotations, &annToken{Annotation: ann, Token: tok})
			} else if reportRefs {
				log15.Warn("spans not match", "repo", repoPath, "rev", repoRev.CommitID, "path", path, "at", tok.Offset, "line", tok.Line, "ident", tok.Text)
			}
		} else if reportRefs {
			log15.Warn("no ref for", "repo", repoPath, "rev", repoRev.CommitID, "path", path, "at", tok.Offset, "line", tok.Line, "ident", tok.Text)
		}
	}
	errors := tt.Errors()
	if len(errors) > 0 {
		log15.Warn("parse errors", "repo", repoPath, "path", path, "errors", errors)
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
		annInfo, err := parseAnnotationURL(u)
		if err != nil {
			return nil, err
		}
		annRepoRev, err := cache.getResolvedRev(cl, ctx, annInfo.RepoRev)
		if err != nil || annRepoRev.CommitID == "" {
			// The ref cannot be resolved to a def (e.g. the def repo doesn't exist);
			// this is a normal condition for the coverage script so swallow the error and continue.
			continue
		}

		defIdx := cache.fetchAndIndexDefs(cl, ctx, &annRepoRev, annInfo.RepoRev.Repo)
		if defIdx == nil {
			continue
		}
		annDefSpec := sourcegraph.DefSpec{
			Repo:     annRepoRev.Repo,
			CommitID: annRepoRev.CommitID,
			UnitType: annInfo.UnitType,
			Unit:     annInfo.Unit,
			Path:     annInfo.Path,
		}
		if def := defIdx.get(annDefSpec); def != nil {
			fileCvg.Defs += 1
		} else {
			if reportDefs {
				log15.Warn("no def", "repo", repoPath, "rev", repoRev.CommitID, "path", path, "at", ann.StartByte, "line", tok.Line, "ident", tok.Text, "key", u)
			}
		}

	}

	if fileCvg.Refs == 0 && reportEmpty {
		log15.Warn("uncovered file", "repo", repoPath, "rev", repoRev.CommitID, "path", path)
	}

	return fileCvg, nil
}

// getCoverage computes coverage data for the given repository
func getCoverage(cl *sourcegraph.Client, ctx context.Context, repoPath, lang string, dryRun, progress, reportRefs, reportDefs, reportEmpty bool) (*srclibRepoCoverage, error) {
	if err := ensureRepoExists(cl, ctx, repoPath); err != nil {
		return nil, err
	}

	start := time.Now()

	repoCvg := srclibRepoCoverage{Repo: repoPath, Day: start.Format("01-02"), Language: lang}
	repoRevSpec, err := cache.getResolvedRev(cl, ctx, routevar.RepoRev{Repo: repoPath})
	if err != nil {
		return nil, err
	}

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
			p.Acquire()
			go func() {
				defer p.Release()
				if progress {
					log15.Info("processing", path, lang)
				}
				fileCvg, err := getFileCoverage(cl, ctx, &repoRevSpec, repoPath, path, lang, reportRefs, reportDefs, reportEmpty)
				if err != nil {
					p.Error(err)
					return
				}
				if fileCvg != nil {
					// fileCvg may be nil for files which are ignored / not indexed
					repoCvgMu.Lock()
					repoCvg.Files = append(repoCvg.Files, fileCvg)
					repoCvgMu.Unlock()
				}
			}()
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
		log15.Warn("missing srclib data version", "repo", repoPath, "rev", repoRevSpec.CommitID)
	}

	repoCvg.Summary = &srclibFileCoverage{}
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

	covJSON, err := json.Marshal([]srclibRepoCoverage{repoCvg})
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
		log15.Error("save coverage stats", "err", err)
		return nil, err
	}

	return &repoCvg, nil
}

func ensureRepoExists(cl *sourcegraph.Client, ctx context.Context, repo string) error {
	// Resolve repo path, and create local mirror for remote repo if needed.
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repo, Remote: true})
	if grpc.Code(err) == codes.NotFound {
		return nil
	} else if err != nil {
		return err
	}

	if remoteRepo := res.RemoteRepo; remoteRepo != nil {
		if actualURI := githubutil.RepoURI(remoteRepo.Owner, remoteRepo.Name); actualURI != repo {
			// Repo path is invalid, possibly because repo has been renamed.
			return fmt.Errorf("repo %s redirects to %s; update dashboard with correct repo path", repo, actualURI)
		}

		// Automatically create a local mirror.
		log15.Info("Creating a local mirror of remote repo", "cloneURL", remoteRepo.HTTPCloneURL)
		_, err := cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_Origin{Origin: remoteRepo.Origin},
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
