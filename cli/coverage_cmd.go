package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage/tokenizer"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/slack"
)

var sgSlackCoverageWebhookURL = os.Getenv("SG_SLACK_COVERAGE_WEBHOOK_URL")

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
	LangpAddr    string `long:"addr" description:"Sourcegraph Language Processor server address" default:"http://localhost:4141"`
	LSPAddr      string `long:"lsp-addr" description:"Sourcegraph LSP server address, only used if -addr is not set"`
	LSPRootPath  string `long:"lsp-root" description:"Root directory location at which LSP server should read files"`
	LSPCaps      bool   `long:"lsp-caps" description:"print the capabilities of the LSP server and exit"`
	Methods      string `long:"methods" description:"API methods to test (definition,hover)" default:"definition,hover"`
	Repo         string `long:"repo" description:"specific repo URI to use"`
	Lang         string `long:"lang" description:"specific coverage language"`
	File         string `long:"file" description:"specific repository filename"`
	Limit        int    `long:"limit" description:"max number of repos to run coverage for"`
	RepoRate     int    `long:"repo-rate" description:"rate at which repositories are concurrently calculated" default:"1"`
	FileRate     int    `long:"file-rate" description:"rate at which files are concurrently calculated" default:"1"`
	TokenRate    int    `long:"rate" description:"rate at which tokens are concurrently calculated" default:"1"`
	Debug        bool   `long:"debug" description:"trace requests which result in an error"`
	AbortOnError int    `long:"abort-on-error" description:"abort on nth request error" default:"-1"`

	backend coverage.Client
	cl      *sourcegraph.Client
}

// langCoverage contains the coverage data for an entire language.
type langCoverage struct {
	Language   string                   // The actual language name, like "Go" or "Java".
	Start, End time.Time                // When coverage calculations began and ended.
	Repos      map[string]*repoCoverage // Coverage data for all repos of this language.
}

// repoCoverage contains the coverage data of a single repository.
type repoCoverage struct {
	RepoURI    string                   // The actual repository URI.
	Start, End time.Time                // When coverage calculations began and ended.
	Files      map[string]*fileCoverage // Coverage data for all files in the repository.

	// Tokens is the total number of tokens across all files.
	Tokens int

	// Number of go-to-defs and hover responses, the sum of these variables
	// will always equal len(Tokens) when 100% coverage is achieved.
	Defs, Hovers           int
	DefErrors, HoverErrors int

	// Distance between token ranges in the file (a measure of accuracy, where
	// zero is perfect, one implies a potential off-by-one error, etc). The
	// larger the number, the worse the result is.
	DefDists []int

	// API request durations.
	DefDurations, HoverDurations []time.Duration
}

func (c *repoCoverage) calcSummary(methods string) {
	for _, f := range c.Files {
		f.calcSummary(methods)
		c.Tokens += len(f.Tokens)
		c.Defs += f.Defs
		c.Hovers += f.Hovers
		c.DefErrors += f.DefErrors
		c.HoverErrors += f.HoverErrors
		c.DefDists = append(c.DefDists, f.DefDists...)
		c.DefDurations = append(c.DefDurations, f.DefDurations...)
		c.HoverDurations = append(c.HoverDurations, f.HoverDurations...)
	}
}

// fileCoverage contains the coverage data of a single file.
type fileCoverage struct {
	Filename   string    // The actual filename.
	Start, End time.Time // When coverage calculations began and ended.
	Tokens     []tokCoverage

	// Number of go-to-defs and hover responses, the sum of these variables
	// will always equal len(Tokens) when 100% coverage is achieved.
	Defs, Hovers           int
	DefErrors, HoverErrors int

	// Distance between token ranges in the file (a measure of accuracy, where
	// zero is perfect, one implies a potential off-by-one error, etc). The
	// larger the number, the worse the result is.
	DefDists []int

	// API request durations.
	DefDurations, HoverDurations []time.Duration
}

func (c *fileCoverage) calcSummary(methods string) {
	for _, tok := range c.Tokens {
		if tok.DefError != nil {
			c.DefErrors++
		} else if strings.Contains(methods, "definition") {
			c.Defs++
			c.DefDurations = append(c.DefDurations, tok.DefEnd.Sub(tok.DefStart))
			c.DefDists = append(c.DefDists, tok.DefDist)
		}
		if tok.HoverError != nil {
			c.HoverErrors++
		} else if strings.Contains(methods, "hover") {
			c.Hovers++
			c.HoverDurations = append(c.HoverDurations, tok.HoverEnd.Sub(tok.HoverStart))
		}
	}
}

type tokCoverage struct {
	// When coverage calculations began and ended (inclusive of all API
	// requests, of which there are multiple per token).
	Start, End time.Time
	Token      tokenizer.Token

	// Error that occured fetching the token as a def or hover.
	DefError, HoverError error

	// Distance between returned ranges (a measure of accuracy where 0 is
	// perfect accuracy).
	DefDist int

	// Exact time at which API requests started and ended.
	DefStart, HoverStart time.Time
	DefEnd, HoverEnd     time.Time
}

func (c *coverageCmd) Execute(args []string) error {
	if c.Lang == "" {
		return fmt.Errorf("must specify exactly one language")
	}
	if c.RepoRate < 1 {
		return fmt.Errorf("repo rate must be at least one")
	}
	if c.FileRate < 1 {
		return fmt.Errorf("file rate must be at least one")
	}

	// Create the connection to LSP server.
	c.cl = cliClient
	var err error
	switch {
	case c.LangpAddr != "":
		c.backend, err = coverage.LangpClient(c.LangpAddr)
	case c.LSPAddr != "":
		c.backend, err = coverage.LSPClient(c.LSPAddr, c.LSPRootPath, c.LSPCaps)
	default:
		return fmt.Errorf("must specify one of --addr or --lsp-addr")
	}
	if err != nil {
		return err
	}
	defer func() {
		if err := c.backend.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	c.logf("Running coverage for %s", c.Lang)
	langCov, err := c.calcLangCoverage(cliContext, c.Lang)
	if err != nil {
		return err
	}
	c.logf("Completed coverage for %s (duration: %s)", langCov.Language, langCov.End.Sub(langCov.Start))

	for _, repo := range langCov.Repos {
		repo.calcSummary(c.Methods)
		c.logf("")
		c.logf("%s finished in %s", repo.RepoURI, repo.End.Sub(repo.Start))
		c.logf("\t%v files", len(repo.Files))
		c.logf("\t%v tokens", repo.Tokens)
		c.logf("")
		c.logf("\tdefinition: %d requests | %d errors", repo.Defs+repo.DefErrors, repo.DefErrors)
		c.logf("\t\tavg. distance: %v", coverage.AvgInt(repo.DefDists))
		c.logf("\t\tavg. response time: %v", coverage.AvgDuration(repo.DefDurations))
		c.logf("\t\tP25 def time:       %v", coverage.Percentile(25, repo.DefDurations))
		c.logf("\t\tP50 def time:       %v", coverage.Percentile(50, repo.DefDurations))
		c.logf("\t\tP75 def time:       %v", coverage.Percentile(75, repo.DefDurations))
		c.logf("\t\tP95 def time:       %v", coverage.Percentile(95, repo.DefDurations))
		c.logf("")
		c.logf("\thover: %d requests | %d errors", repo.Hovers+repo.HoverErrors, repo.HoverErrors)
		c.logf("\t\tavg. response time: %v", coverage.AvgDuration(repo.HoverDurations))
		c.logf("\t\tP25 hover time:     %v", coverage.Percentile(25, repo.HoverDurations))
		c.logf("\t\tP50 hover time:     %v", coverage.Percentile(50, repo.HoverDurations))
		c.logf("\t\tP75 hover time:     %v", coverage.Percentile(75, repo.HoverDurations))
		c.logf("\t\tP95 hover time:     %v", coverage.Percentile(95, repo.HoverDurations))
		c.logf("")

		if c.Debug {
			for _, file := range repo.Files {
				for _, tok := range file.Tokens {
					if tok.DefError != nil {
						c.logf("definition error: %v", tok.DefError)
					}
					if tok.HoverError != nil {
						c.logf("hover error: %v", tok.HoverError)
					}
				}
			}
		}
	}
	return err
}

func (c *coverageCmd) logf(fmtStr string, args ...interface{}) {
	log.Printf(fmtStr, args...)

	if sgSlackCoverageWebhookURL != "" {
		slack.PostMessage(slack.PostOpts{
			Msg:        fmt.Sprintf(fmtStr, args...),
			IconEmoji:  ":chart_with_upwards_trend:",
			WebhookURL: sgSlackCoverageWebhookURL,
		})
	}
}

// calcLangCoverage calculates coverage for the specified languages.
func (c *coverageCmd) calcLangCoverage(ctx context.Context, lang string) (*langCoverage, error) {
	var (
		covMu sync.Mutex
		cov   = &langCoverage{
			Language: lang,
			Start:    time.Now(),
			Repos:    make(map[string]*repoCoverage),
		}
	)

	// Build a list of repo URIs that we will compute coverage for.
	var repos []string
	if specificRepo := c.Repo; specificRepo != "" {
		repos = []string{specificRepo}
	} else {
		repos = langRepos[lang]
	}
	if c.Limit > 0 && len(repos) > c.Limit {
		repos = repos[:c.Limit]
	}

	p := parallel.NewRun(c.RepoRate)
	for _, repo := range repos {
		repo := repo
		p.Acquire()
		go func() {
			defer p.Release()

			repoCov, err := c.calcRepoCoverage(ctx, lang, repo)
			if err != nil {
				p.Error(err)
				return
			}

			covMu.Lock()
			cov.Repos[repo] = repoCov
			covMu.Unlock()
		}()
	}
	err := p.Wait()
	cov.End = time.Now()
	return cov, err
}

// calcRepoCoverage calculates coverage for the specified repo.
func (c *coverageCmd) calcRepoCoverage(ctx context.Context, lang, repoURI string) (*repoCoverage, error) {
	var (
		covMu sync.Mutex
		cov   = &repoCoverage{
			RepoURI: repoURI,
			Start:   time.Now(),
			Files:   make(map[string]*fileCoverage),
		}
	)

	// Ensure the repo exists locally.
	err := c.ensureLocalRepoExists(cliContext, repoURI)
	if err != nil {
		return nil, err
	}

	// Get repository revision spec.
	repoRevSpec, err := c.getRepoRevSpec(ctx, repoURI)
	if err != nil {
		return nil, err
	}

	// Determine which files to calculate coverage with.
	var files []string
	if c.File != "" {
		files = []string{c.File}
	} else {
		// Query a list of the repository files.
		tree, err := c.cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{
			Rev: *repoRevSpec,
		})
		if err != nil {
			return nil, err
		}
		files = tree.Files
	}

	p := parallel.NewRun(c.FileRate)
	for _, file := range files {
		file := file
		p.Acquire()
		go func() {
			defer p.Release()

			fileCov, err := c.calcFileCoverage(ctx, lang, repoURI, repoRevSpec, file)
			if err != nil {
				p.Error(err)
				return
			}
			if fileCov != nil {
				covMu.Lock()
				cov.Files[file] = fileCov
				covMu.Unlock()
			}
		}()
	}
	err = p.Wait()
	cov.End = time.Now()
	return cov, err
}

// calcFileCoverage calculates coverage for the specified file. It may return
// nil if the given file has no tokenizer (i.e. for file types Sourcegraph does
// not support).
func (c *coverageCmd) calcFileCoverage(ctx context.Context, lang, repoURI string, repoRev *sourcegraph.RepoRevSpec, file string) (*fileCoverage, error) {
	var (
		covMu sync.Mutex
		cov   = &fileCoverage{
			Filename: file,
			Start:    time.Now(),
		}
	)

	// Read the entire file contents.
	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: *repoRev,
		Path:    file,
	}
	treeGetOp := sourcegraph.RepoTreeGetOptions{}
	treeGetOp.EntireFile = true
	entry, err := c.cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: entrySpec,
		Opt:   &treeGetOp,
	})
	if err != nil {
		return nil, err
	}

	// TODO: why *interface{} why??
	tx := tokenizer.Lookup(lang, file)
	if tx == nil {
		log.Printf("Warning: no tokenizer for language %q file %q\n", lang, file)
		return nil, nil
	}
	t := *tx
	t.Init(entry.Contents)
	defer t.Done()

	// TODO: concurrency rate here
	p := parallel.NewRun(c.TokenRate)
	for {
		tok := t.Next()
		if tok == nil {
			break
		}

		p.Acquire()
		go func() {
			defer p.Release()

			tokCov, err := c.calcTokCoverage(ctx, lang, repoURI, repoRev, file, tok)
			if err != nil {
				p.Error(err)
				return
			}
			covMu.Lock()
			cov.Tokens = append(cov.Tokens, *tokCov)
			covMu.Unlock()
		}()
	}
	if errs := t.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("tokenizer: error: %s", strings.Join(errs, ", "))
	}
	if err := p.Wait(); err != nil {
		return nil, err
	}

	cov.End = time.Now()
	return cov, nil
}

func (c *coverageCmd) calcTokCoverage(ctx context.Context, lang, repoURI string, repoRev *sourcegraph.RepoRevSpec, fileURI string, tok *tokenizer.Token) (*tokCoverage, error) {
	var cov = &tokCoverage{
		Token: *tok,
		Start: time.Now(),
	}

	p := &langp.Position{
		Repo:      repoURI,
		Commit:    repoRev.CommitID,
		File:      fileURI,
		Line:      tok.Line - 1,   // Subtract one because tokenizer is one-based
		Character: tok.Column - 1, // Subtract one because tokenizer is one-based
	}

	if strings.Contains(c.Methods, "definition") {
		cov.DefStart = time.Now()
		def, err := c.backend.Definition(p)
		cov.DefEnd = time.Now()
		if err != nil {
			cov.DefError = err
			coverage.Abort(c.Debug, c.AbortOnError, err, coverage.ErrorCase{
				Method:    "definition",
				Repo:      p.Repo,
				File:      p.File,
				Text:      tok.Text,
				Line:      p.Line,
				Character: p.Character,
			})
		} else {
			cov.DefDist = coverage.Dist(def.LSP(), coverage.TokRange(*tok))
		}
	}

	if strings.Contains(c.Methods, "hover") {
		cov.HoverStart = time.Now()
		_, err := c.backend.Hover(p)
		cov.HoverEnd = time.Now()
		if err != nil {
			cov.HoverError = err
			coverage.Abort(c.Debug, c.AbortOnError, err, coverage.ErrorCase{
				Method:    "hover",
				Repo:      p.Repo,
				File:      p.File,
				Text:      tok.Text,
				Line:      p.Line,
				Character: p.Character,
			})
		}
	}

	// TODO: local refs

	cov.End = time.Now()
	return cov, nil
}

func (c *coverageCmd) getRepoRevSpec(ctx context.Context, repoURI string) (*sourcegraph.RepoRevSpec, error) {
	// Resolve the repo.
	res, err := c.cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoURI})
	if err != nil {
		return nil, err
	}

	// Get the repo.
	repo, err := c.cl.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: res.Repo})
	if err != nil {
		return nil, err
	}

	// Resolve the default branch revision.
	resRev, err := c.cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: res.Repo,
		Rev:  repo.DefaultBranch,
	})
	if err != nil {
		return nil, err
	}

	return &sourcegraph.RepoRevSpec{
		Repo:     res.Repo,
		CommitID: resRev.CommitID,
	}, nil
}

// ensureLocalRepoExists ensures that a repository with the specified URI
// exists locally. If the repo does not exist locally, the remote gRPC endpoint
// is consulted to clone the repository.
func (c *coverageCmd) ensureLocalRepoExists(ctx context.Context, repo string) error {
	res, err := c.cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repo, Remote: true})
	if grpc.Code(err) == codes.NotFound {
		// Repo doesn't exist on the remote, but maybe we have it locally
		// already.
		return nil
	} else if err != nil {
		return err
	}

	if res.RemoteRepo == nil {
		// No remote repo, we can't clone it.
		return nil
	}

	actualURI := githubutil.RepoURI(res.RemoteRepo.Owner, res.RemoteRepo.Name)
	if actualURI != repo {
		// Repo path is invalid, possibly because repo has been renamed.
		return fmt.Errorf("repo %s redirects to %s; update to correct repo path", repo, actualURI)
	}

	// Automatically create a local mirror.
	log15.Info("Creating a local mirror of remote repo", "cloneURL", res.RemoteRepo.HTTPCloneURL)
	_, err = c.cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_Origin{Origin: res.RemoteRepo.Origin},
	})
	return err
}

// tokRange returns an LSP equivilent range for the given token.
func (c *coverageCmd) tokRange(t tokenizer.Token) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      t.Line,
			Character: t.Column,
		},
		End: lsp.Position{
			Line:      t.Line,
			Character: t.Column + len(t.Text),
		},
	}
}

// rangeDist calculates the distance between the two ranges. A distance
// of 0 implies the two ranges are equal, a distance of 1 implies there may be
// an off-by-one error, etc. Aside from "larger distances mean the two are less
// equal" the definition of distance is not strictly defined (except by the
// source).
func (c *coverageCmd) rangeDist(x, y lsp.Range) int {
	// dist calculates the distance between the two integers.
	dist := func(x, y int) int {
		v := x - y
		if v < 0 {
			return -v
		}
		return v
	}

	// We could return start character, start line, end character, and end line
	// distances all independently for better introspection, but this is good
	// enough until proven otherwise.
	d := dist(x.Start.Character, y.Start.Character)
	d += dist(x.Start.Line, y.Start.Line)
	d += dist(x.End.Character, y.End.Character)
	d += dist(x.End.Line, y.End.Line)
	return d
}
