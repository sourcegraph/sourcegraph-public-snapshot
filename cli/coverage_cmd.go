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

	cssParser "github.com/chris-ramon/douceur/parser"
	"github.com/gorilla/mux"
	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
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
	}
}

type coverageCmd struct {
	Repo    string `long:"repo" description:"repo URI"`
	Lang    string `long:"lang" description:"coverage language"`
	Refresh bool   `long:"refresh" description:"refresh the coverage information or compute it if it doesn't exist yet"`
}

// fileCoverage contains coverage data for a single file
type fileCoverage struct {
	Path     string // the file path
	Language string // the language name
	Idents   int    // # of identifiers in the file
	Refs     int    // # of refs in the file (i.e. annotations)
	Defs     int    // # of annotations (URLs) which resolve to real defs
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
	Summary        []*fileCoverage // summation over Files, per language
	SrclibVersions []*srclibVersion
	Day            string
	Duration       float64 // time to compute coverage (in seconds)
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

	// FetchDefsMu allows at most one goroutine to fetch defs;
	// otherwise multiple goroutines may make calls to fetch defs
	// for a given repository and overwork the server.
	FetchDefsMu sync.Mutex
}

// cache is a global instance of coverageCache
var cache *coverageCache

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
func (c *coverageCache) getResolvedRev(cl *sourcegraph.Client, ctx context.Context, repoSpec *sourcegraph.RepoSpec) (string, error) {
	c.ResolvedRevMu.Lock()
	defer c.ResolvedRevMu.Unlock()

	if commitID, ok := c.ResolvedRevCache[repoSpec.URI]; ok {
		return commitID, nil
	}

	repo, err := cl.Repos.Get(ctx, repoSpec)
	if err != nil {
		return "", err
	}

	res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: *repoSpec, Rev: repo.DefaultBranch})
	if err != nil {
		return "", err
	}

	c.ResolvedRevCache[repoSpec.URI] = res.CommitID
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

	c.FetchDefsMu.Lock()
	defer c.FetchDefsMu.Unlock()

	rr := repoRevKey(repoRev)
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
	return fmt.Sprintf("%s@%s", repoRev.RepoSpec.URI, repoRev.CommitID)
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
				cov, err := getCoverage(cl, cliContext, repo, lang)
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
				RepoSpec: repoRev.RepoSpec,
				CommitID: repoRev.Rev,
			}, &sourcegraph.DefSpec{
				Repo:     def.RepoSpec.URI,
				CommitID: def.Rev,
				UnitType: def.UnitType,
				Unit:     def.Unit,
				Path:     def.Path,
			}, nil
	} else {
		return nil, nil, fmt.Errorf("error parsing mux vars for annotation url %s", annUrl)
	}
}

// getFileCoverage computes the coverage data for a single file in a repository
func getFileCoverage(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, path, lang string) (*fileCoverage, error) {
	fileCvg := &fileCoverage{Path: path, Language: lang}

	// TODO(rothfels): add other language support.
	switch lang {
	case "Go":
		if !strings.HasSuffix(path, ".go") {
			return nil, nil
		}
	case "CSS":
		if !strings.HasSuffix(path, ".css") {
			return nil, nil
		}
		return getCSSFileCoverage(cl, ctx, repoRev, path, lang)
	default:
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
			annsByStartByte[ann.StartByte+1] = ann // off by one
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
			return nil, err
		}
		if annRepoRev.CommitID == "" {
			commitID, err := cache.getResolvedRev(cl, ctx, &annRepoRev.RepoSpec)
			if err != nil || commitID == "" {
				// The ref cannot be resolved to a def (e.g. the def repo doesn't exist);
				// this is a normal condition for the coverage script so swallow the error and continue.
				continue
			}
			annRepoRev.CommitID = commitID
		}

		defIdx := cache.fetchAndIndexDefs(cl, ctx, annRepoRev)
		if defIdx == nil {
			continue
		}
		if def := defIdx.get(defKey(annDefSpec)); def != nil {
			fileCvg.Defs += 1
		}
	}

	log15.Info("computed coverage", "path", path, "idents", fileCvg.Idents, "refs", fileCvg.Refs, "defs", fileCvg.Defs)
	return fileCvg, nil
}

// getCoverage computes coverage data for the given repository
func getCoverage(cl *sourcegraph.Client, ctx context.Context, repoURI, lang string) (*repoCoverage, error) {
	if err := ensureRepoExists(cl, ctx, repoURI); err != nil {
		return nil, err
	}

	start := time.Now()

	repoSpec := sourcegraph.RepoSpec{URI: repoURI}
	commitID, err := cache.getResolvedRev(cl, ctx, &repoSpec)
	if err != nil {
		return nil, err
	}
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, CommitID: commitID}

	tree, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: repoRevSpec})
	if err != nil {
		return nil, err
	}

	p := parallel.NewRun(10)
	var repoCvgMu sync.Mutex
	repoCvg := repoCoverage{Repo: repoURI, Day: start.Format("01-02")}
	for _, path := range tree.Files {
		path := path
		p.Do(func() error {
			fileCvg, err := getFileCoverage(cl, ctx, &repoRevSpec, path, lang)
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

func getCSSFileCoverage(cl *sourcegraph.Client, ctx context.Context, repoRev *sourcegraph.RepoRevSpec, path, lang string) (*fileCoverage, error) {
	fileCvg := &fileCoverage{Path: path, Language: lang}
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
	data := string(entry.Contents)
	stylesheet, err := cssParser.Parse(data)
	if err != nil {
		return nil, err
	}
	for _, r := range stylesheet.Rules {
		for _, s := range r.Selectors {
			defStart, _ := findOffsets(data, s.Line, s.Column, s.Value)
			if defStart == 0 {
				defStart = 1
			}
			fileCvg.Idents += 1
			if _, ok := annsByStartByte[uint32(defStart)]; ok {
				fileCvg.Refs += 1
			}
		}
		for _, d := range r.Declarations {
			fileCvg.Idents += 1
			s, _ := findOffsets(data, d.Line, d.Column, d.Property)
			if _, ok := annsByStartByte[uint32(s)]; ok {
				fileCvg.Refs += 1
			}
		}
	}
	for _, ann := range anns.Annotations {
		if isExternalAnnURL(*ann) {
			fileCvg.Refs += 1
		}
	}
	log15.Info("computed CSS coverage", "path", path, "idents", fileCvg.Idents, "refs", fileCvg.Refs, "defs", fileCvg.Defs)
	return fileCvg, nil
}

// isExternalLinkAnn checks if given url is an external URL.
func isExternalAnnURL(ann sourcegraph.Annotation) bool {
	if strings.HasPrefix(ann.URL, "https://developer.mozilla.org") || strings.HasPrefix(ann.URL, "http://developer.mozilla.org") {
		return true
	}
	return false
}

// TODO (chris): Replace `findOffsets` with an array of byte/rune offsets lookup strategy.
// See: https://github.com/sourcegraph/srclib-css/pull/1#discussion_r61206972
// findOffsets discovers the start & end offset of given token on fileText, uses the given line & column as input
// to discover the start offset which is used to calculate the end offset.
// Returns (-1, -1) if offsets were not found.
func findOffsets(fileText string, line, column int, token string) (start, end int) {

	// we count our current line and column position.
	currentCol := 1
	currentLine := 1

	for offset, ch := range fileText {
		// see if we found where we wanted to go to.
		if currentLine == line && currentCol == column {
			end = offset + len([]byte(token))
			return offset, end
		}

		// line break - increment the line counter and reset the column.
		if ch == '\n' {
			currentLine++
			currentCol = 1
		} else {
			currentCol++
		}
	}

	return -1, -1 // not found.
}
