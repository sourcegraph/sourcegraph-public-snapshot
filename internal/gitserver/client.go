package gitserver

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/grafana/regexp"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/go-rendezvous"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/migration"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const git = "git"

var (
	clientFactory  = httpcli.NewInternalClientFactory("gitserver")
	defaultDoer, _ = clientFactory.Doer()
	defaultLimiter = parallel.NewRun(500)
)

var ClientMocks, emptyClientMocks struct {
	GetObject               func(repo api.RepoName, objectName string) (*gitdomain.GitObject, error)
	Archive                 func(ctx context.Context, repo api.RepoName, opt ArchiveOptions) (_ io.ReadCloser, err error)
	LocalGitserver          bool
	LocalGitCommandReposDir string
}

// ResetClientMocks clears the mock functions set on Mocks (so that subsequent
// tests don't inadvertently use them).
func ResetClientMocks() {
	ClientMocks = emptyClientMocks
}

var _ Client = &clientImplementor{}

// NewClient returns a new gitserver.Client.
func NewClient(db database.DB) Client {
	return &clientImplementor{
		logger: sglog.Scoped("NewClient", "returns a new gitserver.Client"),
		addrs: func() []string {
			return conf.Get().ServiceConnections().GitServers
		},
		pinned:      pinnedReposFromConfig,
		db:          db,
		httpClient:  defaultDoer,
		HTTPLimiter: defaultLimiter,
		// Use the binary name for userAgent. This should effectively identify
		// which service is making the request (excluding requests proxied via the
		// frontend internal API)
		userAgent:  filepath.Base(os.Args[0]),
		operations: getOperations(),
	}
}

// NewTestClient returns a test client that will use the given hard coded list of
// addresses instead of reading them from config.
func NewTestClient(cli httpcli.Doer, db database.DB, addrs []string) Client {
	return &clientImplementor{
		logger: sglog.Scoped("NewTestClient", "Test New client"),
		addrs: func() []string {
			return addrs
		},
		pinned:      pinnedReposFromConfig,
		httpClient:  cli,
		HTTPLimiter: parallel.NewRun(500),
		// Use the binary name for userAgent. This should effectively identify
		// which service is making the request (excluding requests proxied via the
		// frontend internal API)
		userAgent:  filepath.Base(os.Args[0]),
		db:         db,
		operations: newOperations(&observation.TestContext),
	}
}

// clientImplementor is a gitserver client.
type clientImplementor struct {
	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run

	// userAgent is a string identifying who the client is. It will be logged in
	// the telemetry in gitserver.
	userAgent string

	// HTTP client to use
	httpClient httpcli.Doer

	// logger is used for all logging and logger creation
	logger sglog.Logger

	// addrs is a function which should return the addresses for gitservers. It
	// is called each time a request is made. The function must be safe for
	// concurrent use. It may return different results at different times.
	addrs func() []string

	// pinned holds a map of repositories(key) pinned to a particular gitserver instance(value). This function
	// should query the conf to fetch a fresh map of pinned repos, so that we don't have to proactively watch for conf changes
	// and sync the pinned map.
	pinned func() map[string]string

	// db is a connection to the database
	db database.DB

	// operations are used for internal observability
	operations *operations
}

type RawBatchLogResult struct {
	Stdout string
	Error  error
}
type BatchLogCallback func(repoCommit api.RepoCommit, gitLogResult RawBatchLogResult) error

type Client interface {
	// AddrForRepo returns the gitserver address to use for the given repo name.
	AddrForRepo(context.Context, api.RepoName) (string, error)

	// ArchiveReader streams back the file contents of an archived git repo.
	ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options ArchiveOptions) (io.ReadCloser, error)

	// BatchLog invokes the given callback with the `git log` output for a batch of repository
	// and commit pairs. If the invoked callback returns a non-nil error, the operation will begin
	// to abort processing further results.
	BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) error

	// BlameFile returns Git blame information about a file.
	BlameFile(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, path string, opt *BlameOptions) ([]*Hunk, error)

	// CreateCommitFromPatch will attempt to create a commit from a patch
	// If possible, the error returned will be of type protocol.CreateCommitFromPatchError
	CreateCommitFromPatch(context.Context, protocol.CreateCommitFromPatchRequest) (string, error)

	// GetDefaultBranch returns the name of the default branch and the commit it's
	// currently at from the given repository. If short is true, then `main` instead
	// of `refs/heads/main` would be returned.
	//
	// If the repository is empty or currently being cloned, empty values and no
	// error are returned.
	GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error)

	// GetObject fetches git object data in the supplied repo
	GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error)

	// HasCommitAfter indicates the staleness of a repository. It returns a boolean indicating if a repository
	// contains a commit past a specified date.
	HasCommitAfter(ctx context.Context, repo api.RepoName, date string, revspec string, checker authz.SubRepoPermissionChecker) (bool, error)

	// IsRepoCloneable returns nil if the repository is cloneable.
	IsRepoCloneable(context.Context, api.RepoName) error

	// ListRefs returns a list of all refs in the repository.
	ListRefs(ctx context.Context, repo api.RepoName) ([]gitdomain.Ref, error)

	// ListBranches returns a list of all branches in the repository.
	ListBranches(ctx context.Context, repo api.RepoName, opt BranchesOptions) ([]*gitdomain.Branch, error)

	// MergeBase returns the merge base commit for the specified commits.
	MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (api.CommitID, error)

	// P4Exec sends a p4 command with given arguments and returns an io.ReadCloser for the output.
	P4Exec(_ context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error)

	// Remove removes the repository clone from gitserver.
	Remove(context.Context, api.RepoName) error

	// RemoveFrom removes the repository clone from the given gitserver.
	RemoveFrom(ctx context.Context, repo api.RepoName, from string) error

	// RendezvousAddrForRepo returns the gitserver address to use for the given
	// repo name using the Rendezvous hashing scheme.
	RendezvousAddrForRepo(api.RepoName) string

	RepoCloneProgress(context.Context, ...api.RepoName) (*protocol.RepoCloneProgressResponse, error)

	// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
	// used.
	//
	// Error cases:
	// * Repo does not exist: gitdomain.RepoNotExistError
	// * Commit does not exist: gitdomain.RevisionNotFoundError
	// * Empty repository: gitdomain.RevisionNotFoundError
	// * Other unexpected errors.
	ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (api.CommitID, error)

	// ResolveRevisions expands a set of RevisionSpecifiers (which may include hashes, globs, refs, or glob exclusions)
	// into an equivalent set of commit hashes
	ResolveRevisions(_ context.Context, repo api.RepoName, _ []protocol.RevisionSpecifier) ([]string, error)

	// ReposStats will return a map of the ReposStats for each gitserver in a
	// map. If we fail to fetch a stat from a gitserver, it won't be in the
	// returned map and will be appended to the error. If no errors occur err will
	// be nil.
	//
	// Note: If the statistics for a gitserver have not been computed, the
	// UpdatedAt field will be zero. This can happen for new gitservers.
	ReposStats(context.Context) (map[string]*protocol.ReposStats, error)

	// RequestRepoMigrate is effectively RequestRepoUpdate but with some additional metadata to make
	// gitserver instances clone a repo from one instance to another
	RequestRepoMigrate(ctx context.Context, repo api.RepoName, from, to string) (*protocol.RepoUpdateResponse, error)

	// RequestRepoUpdate is the new protocol endpoint for synchronous requests
	// with more detailed responses. Do not use this if you are not repo-updater.
	//
	// Repo updates are not guaranteed to occur. If a repo has been updated
	// recently (within the Since duration specified in the request), the
	// update won't happen.
	RequestRepoUpdate(context.Context, api.RepoName, time.Duration) (*protocol.RepoUpdateResponse, error)

	// Search executes a search as specified by args, streaming the results as
	// it goes by calling onMatches with each set of results it receives in
	// response.
	Search(_ context.Context, _ *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, _ error)

	// Stat returns a FileInfo describing the named file at commit.
	Stat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error)

	// DiffPath returns a position-ordered slice of changes (additions or deletions)
	// of the given path between the given source and target commits.
	DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error)

	// ReadDir reads the contents of the named directory at commit.
	ReadDir(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error)

	// NewFileReader returns an io.ReadCloser reading from the named file at commit.
	// The caller should always close the reader after use
	NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) (io.ReadCloser, error)

	// DiffSymbols performs a diff command which is expected to be parsed by our symbols package
	DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error)

	// ListFiles returns a list of root-relative file paths matching the given
	// pattern in a particular commit of a repository.
	ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) ([]string, error)

	// Commits returns all commits matching the options.
	Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error)

	// FirstEverCommit returns the first commit ever made to the repository.
	FirstEverCommit(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error)

	// ListTags returns a list of all tags in the repository. If commitObjs is non-empty, only all tags pointing at those commits are returned.
	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) ([]*gitdomain.Tag, error)

	// ListDirectoryChildren fetches the list of children under the given directory
	// names. The result is a map keyed by the directory names with the list of files
	// under each.
	ListDirectoryChildren(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, dirnames []string) (map[string][]string, error)

	// Diff returns an iterator that can be used to access the diff between two
	// commits on a per-file basis. The iterator must be closed with Close when no
	// longer required.
	Diff(ctx context.Context, opts DiffOptions, checker authz.SubRepoPermissionChecker) (*DiffFileIterator, error)

	// ReadFile returns the first maxBytes of the named file at commit. If maxBytes <= 0, the entire
	// file is read. (If you just need to check a file's existence, use Stat, not ReadFile.)
	ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string, checker authz.SubRepoPermissionChecker) ([]byte, error)

	// BranchesContaining returns a map from branch names to branch tip hashes for
	// each branch containing the given commit.
	BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) ([]string, error)

	// RefDescriptions returns a map from commits to descriptions of the tip of each
	// branch and tag of the given repository.
	RefDescriptions(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker, gitObjs ...string) (map[string][]gitdomain.RefDescription, error)

	// CommitExists determines if the given commit exists in the given repository.
	CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID, checker authz.SubRepoPermissionChecker) (bool, error)

	// CommitsExist determines if the given commits exists in the given repositories. This function returns
	// a slice of the same size as the input slice, true indicating that the commit at the symmetric index
	// exists.
	CommitsExist(ctx context.Context, repoCommits []api.RepoCommit, checker authz.SubRepoPermissionChecker) ([]bool, error)

	// Head determines the tip commit of the default branch for the given repository.
	// If no HEAD revision exists for the given repository (which occurs with empty
	// repositories), a false-valued flag is returned along with a nil error and
	// empty revision.
	Head(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker) (string, bool, error)

	// CommitDate returns the time that the given commit was committed. If the given
	// revision does not exist, a false-valued flag is returned along with a nil
	// error and zero-valued time.
	CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID, checker authz.SubRepoPermissionChecker) (string, time.Time, bool, error)

	// CommitGraph returns the commit graph for the given repository as a mapping
	// from a commit to its parents. If a commit is supplied, the returned graph will
	// be rooted at the given commit. If a non-zero limit is supplied, at most that
	// many commits will be returned.
	CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error)

	// CommitsUniqueToBranch returns a map from commits that exist on a particular
	// branch in the given repository to their committer date. This set of commits is
	// determined by listing `{branchName} ^HEAD`, which is interpreted as: all
	// commits on {branchName} not also on the tip of the default branch. If the
	// supplied branch name is the default branch, then this method instead returns
	// all commits reachable from HEAD.
	CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time, checker authz.SubRepoPermissionChecker) (map[string]time.Time, error)

	// LsFiles returns the output of `git ls-files`
	LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error)

	// GetCommits returns a git commit object describing each of the given repository and commit pairs. This
	// function returns a slice of the same size as the input slice. Values in the output slice may be nil if
	// their associated repository or commit are unresolvable.
	//
	// If ignoreErrors is true, then errors arising from any single failed git log operation will cause the
	// resulting commit to be nil, but not fail the entire operation.
	GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool, checker authz.SubRepoPermissionChecker) ([]*gitdomain.Commit, error)

	// GetCommit returns the commit with the given commit ID, or ErrCommitNotFound if no such commit
	// exists.
	//
	// The remoteURLFunc is called to get the Git remote URL if it's not set in repo and if it is
	// needed. The Git remote URL is only required if the gitserver doesn't already contain a clone of
	// the repository or if the commit must be fetched from the remote.
	GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions, checker authz.SubRepoPermissionChecker) (*gitdomain.Commit, error)

	// GetBehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
	// revspecs).
	GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string) (*gitdomain.BehindAhead, error)

	// ContributorCount returns the number of commits grouped by contributor
	ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) ([]*gitdomain.ContributorCount, error)

	// LogReverseEach runs git log in reverse order and calls the given callback for each entry.
	LogReverseEach(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) error

	// RevList makes a git rev-list call and iterates through the resulting commits, calling the provided
	// onCommit function for each.
	RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (bool, error)) error

	Addrs() []string
}

func (c *clientImplementor) Addrs() []string {
	return c.addrs()
}

func (c *clientImplementor) AddrForRepo(ctx context.Context, repo api.RepoName) (string, error) {
	addrs := c.Addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	return AddrForRepo(ctx, c.userAgent, c.db, repo, GitServerAddresses{
		Addresses:     addrs,
		PinnedServers: c.pinned(),
	})
}

func (c *clientImplementor) RendezvousAddrForRepo(repo api.RepoName) string {
	addrs := c.Addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	if repoPinned, addr := getPinnedRepoAddr(string(repo), c.pinned()); repoPinned {
		return addr
	}
	return RendezvousAddrForRepo(repo, addrs)
}

var addrForRepoInvoked = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_invoked",
	Help: "Number of times gitserver.AddrForRepo was invoked",
}, []string{"user_agent"})

// AddrForRepo returns the gitserver address to use for the given repo name.
// It should never be called with a nil addresses pointer.
func AddrForRepo(ctx context.Context, userAgent string, db database.DB, repo api.RepoName, addresses GitServerAddresses) (string, error) {
	addrForRepoInvoked.WithLabelValues(userAgent).Inc()

	repo = protocol.NormalizeRepo(repo) // in case the caller didn't already normalize it
	rs := string(repo)
	if repoPinned, addr := getPinnedRepoAddr(string(repo), addresses.PinnedServers); repoPinned {
		return addr, nil
	}

	useRendezvous, err := shouldUseRendezvousHashing(ctx, db, rs)
	if err != nil {
		return "", err
	}
	if useRendezvous {
		return RendezvousAddrForRepo(repo, addresses.Addresses), nil
	}

	return addrForKey(rs, addresses.Addresses), nil
}

type GitServerAddresses struct {
	Addresses     []string
	PinnedServers map[string]string
}

// RendezvousAddrForRepo returns the gitserver address to use for the given repo name using the
// Rendezvous hashing scheme.
//
// It should never be called with an empty slice.
func RendezvousAddrForRepo(repo api.RepoName, addrs []string) string {
	r := rendezvous.New(addrs, xxhash.Sum64String)
	return r.Lookup(string(protocol.NormalizeRepo(repo)))
}

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func addrForKey(key string, addrs []string) string {
	sum := md5.Sum([]byte(key))
	serverIndex := binary.BigEndian.Uint64(sum[:]) % uint64(len(addrs))
	return addrs[serverIndex]
}

// ArchiveOptions contains options for the Archive func.
type ArchiveOptions struct {
	Treeish   string               // the tree or commit to produce an archive for
	Format    ArchiveFormat        // format of the resulting archive (usually "tar" or "zip")
	Pathspecs []gitdomain.Pathspec // if nonempty, only include these pathspecs.
}

type BatchLogOptions protocol.BatchLogRequest

func (opts BatchLogOptions) LogFields() []log.Field {
	return []log.Field{
		log.Int("numRepoCommits", len(opts.RepoCommits)),
		log.String("Format", opts.Format),
	}
}

// archiveReader wraps the StdoutReader yielded by gitserver's
// RemoteGitCommand.StdoutReader with one that knows how to report a repository-not-found
// error more carefully.
type archiveReader struct {
	base io.ReadCloser
	repo api.RepoName
	spec string
}

// Read checks the known output behavior of the StdoutReader.
func (a *archiveReader) Read(p []byte) (int, error) {
	n, err := a.base.Read(p)
	if err != nil {
		// handle the special case where git archive failed because of an invalid spec
		if strings.Contains(err.Error(), "Not a valid object") {
			return 0, &gitdomain.RevisionNotFoundError{Repo: a.repo, Spec: a.spec}
		}
	}
	return n, err
}

func (a *archiveReader) Close() error {
	return a.base.Close()
}

// archiveURL returns a URL from which an archive of the given Git repository can
// be downloaded from.
func (c *clientImplementor) archiveURL(ctx context.Context, repo api.RepoName, opt ArchiveOptions) (*url.URL, error) {
	q := url.Values{
		"repo":    {string(repo)},
		"treeish": {opt.Treeish},
		"format":  {string(opt.Format)},
	}

	for _, pathspec := range opt.Pathspecs {
		q.Add("path", string(pathspec))
	}

	addrForRepo, err := c.AddrForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Scheme:   "http",
		Host:     addrForRepo,
		Path:     "/archive",
		RawQuery: q.Encode(),
	}, nil
}

type badRequestError struct{ error }

func (e badRequestError) BadRequest() bool { return true }

func (c *RemoteGitCommand) sendExec(ctx context.Context) (_ io.ReadCloser, _ http.Header, errRes error) {
	repoName := protocol.NormalizeRepo(c.repo)

	span, ctx := ot.StartSpanFromContext(ctx, "Client.sendExec")
	defer func() {
		if errRes != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", errRes.Error())
		}
		span.Finish()
	}()
	span.SetTag("request", "Exec")
	span.SetTag("repo", c.repo)
	span.SetTag("args", c.args[1:])

	// Check that ctx is not expired.
	if err := ctx.Err(); err != nil {
		deadlineExceededCounter.Inc()
		return nil, nil, err
	}

	req := &protocol.ExecRequest{
		Repo:           repoName,
		EnsureRevision: c.EnsureRevision(),
		Args:           c.args[1:],
		NoTimeout:      c.noTimeout,
	}
	resp, err := c.execFn(ctx, repoName, "exec", req)
	if err != nil {
		return nil, nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return resp.Body, resp.Trailer, nil

	case http.StatusNotFound:
		var payload protocol.NotFoundPayload
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, nil, err
		}
		resp.Body.Close()
		return nil, nil, &gitdomain.RepoNotExistError{Repo: repoName, CloneInProgress: payload.CloneInProgress, CloneProgress: payload.CloneProgress}

	default:
		resp.Body.Close()
		return nil, nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *clientImplementor) Search(ctx context.Context, args *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "GitserverClient.Search")
	span.SetTag("repo", string(args.Repo))
	span.SetTag("query", args.Query.String())
	span.SetTag("diff", args.IncludeDiff)
	span.SetTag("limit", args.Limit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	repoName := protocol.NormalizeRepo(args.Repo)

	protocol.RegisterGob()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(args); err != nil {
		return false, err
	}

	addrForRepo, err := c.AddrForRepo(ctx, repoName)
	if err != nil {
		return false, err
	}

	uri := "http://" + addrForRepo + "/search"
	resp, err := c.do(ctx, repoName, "POST", uri, buf.Bytes())
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var (
		decodeErr error
		eventDone protocol.SearchEventDone
	)
	dec := StreamSearchDecoder{
		OnMatches: func(e protocol.SearchEventMatches) {
			onMatches(e)
		},
		OnDone: func(e protocol.SearchEventDone) {
			eventDone = e
		},
		OnUnknown: func(event, _ []byte) {
			decodeErr = errors.Errorf("unknown event %s", event)
		},
	}

	if err := dec.ReadAll(resp.Body); err != nil {
		return false, err
	}

	if decodeErr != nil {
		return false, decodeErr
	}

	return eventDone.LimitHit, eventDone.Err()
}

func (c *clientImplementor) P4Exec(ctx context.Context, host, user, password string, args ...string) (_ io.ReadCloser, _ http.Header, errRes error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Client.P4Exec")
	defer func() {
		if errRes != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", errRes.Error())
		}
		span.Finish()
	}()
	span.SetTag("request", "P4Exec")
	span.SetTag("host", host)
	span.SetTag("args", args)

	// Check that ctx is not expired.
	if err := ctx.Err(); err != nil {
		deadlineExceededCounter.Inc()
		return nil, nil, err
	}

	req := &protocol.P4ExecRequest{
		P4Port:   host,
		P4User:   user,
		P4Passwd: password,
		Args:     args,
	}
	resp, err := c.httpPost(ctx, "", "p4-exec", req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, nil, errors.Errorf("unexpected status code: %d - %s", resp.StatusCode, readResponseBody(resp.Body))
	}

	return resp.Body, resp.Trailer, nil
}

var deadlineExceededCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_gitserver_client_deadline_exceeded",
	Help: "Times that Client.sendExec() returned context.DeadlineExceeded",
})

// BatchLog invokes the given callback with the `git log` output for a batch of repository
// and commit pairs. If the invoked callback returns a non-nil error, the operation will begin
// to abort processing further results.
func (c *clientImplementor) BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) (err error) {
	ctx, _, endObservation := c.operations.batchLog.With(ctx, &err, observation.Args{LogFields: opts.LogFields()})
	defer endObservation(1, observation.Args{})

	// Make a request to a single gitserver shard and feed the results to the user-supplied
	// callback. This function is invoked multiple times (and concurrently) in the loops below
	// this function definition.
	performLogRequestToShard := func(ctx context.Context, addr string, repoCommits []api.RepoCommit) (err error) {
		var numProcessed int
		repoNames := repoNamesFromRepoCommits(repoCommits)

		ctx, logger, endObservation := c.operations.batchLogSingle.With(ctx, &err, observation.Args{
			LogFields: []log.Field{
				log.String("addr", addr),
				log.Int("numRepos", len(repoNames)),
				log.Int("numRepoCommits", len(repoCommits)),
			},
		})
		defer func() {
			endObservation(1, observation.Args{
				LogFields: []log.Field{
					log.Int("numProcessed", numProcessed),
				},
			})
		}()

		uri := "http://" + addr + "/batch-log"
		repoName := api.RepoName(strings.Join(repoNames, ",")) // only used to label spans

		request := protocol.BatchLogRequest{
			RepoCommits: repoCommits,
			Format:      opts.Format,
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(request); err != nil {
			return err
		}

		resp, err := c.do(ctx, repoName, "POST", uri, buf.Bytes())
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		logger.Log(log.Int("resp.StatusCode", resp.StatusCode))

		if resp.StatusCode != http.StatusOK {
			return errors.Newf("http status %d: %s", resp.StatusCode, readResponseBody(io.LimitReader(resp.Body, 200)))
		}

		var response protocol.BatchLogResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}
		logger.Log(log.Int("numResults", len(response.Results)))

		for _, result := range response.Results {
			var err error
			if result.CommandError != "" {
				err = errors.New(result.CommandError)
			}

			rawResult := RawBatchLogResult{
				Stdout: result.CommandOutput,
				Error:  err,
			}
			if err := callback(result.RepoCommit, rawResult); err != nil {
				return errors.Wrap(err, "commitLogCallback")
			}

			numProcessed++
		}

		return nil
	}

	// Construct batches of requests keyed by the address of the server that will receive the batch.
	// The results from gitserver will have to be re-interlaced before returning to the client, so we
	// don't need to be particularly concerned about order here.

	batches := make(map[string][]api.RepoCommit, len(opts.RepoCommits))
	addrsByName := make(map[api.RepoName]string, len(opts.RepoCommits))

	for _, repoCommit := range opts.RepoCommits {
		addr, ok := addrsByName[repoCommit.Repo]
		if !ok {
			addr, err = c.AddrForRepo(ctx, repoCommit.Repo)
			if err != nil {
				return err
			}

			addrsByName[repoCommit.Repo] = addr
		}

		batches[addr] = append(batches[addr], api.RepoCommit{
			Repo:     repoCommit.Repo,
			CommitID: repoCommit.CommitID,
		})
	}

	// Perform each batch request concurrently up to a maximum limit of 32 requests
	// in-flight at one time.
	//
	// This limit will be useless in practice most of the  time as we should only be
	// making one request per shard and instances should _generally_ have fewer than
	// 32 gitserver shards. This condition is really to catch unexpected bad behavior.
	// At the time this limit was chosen, we have 20 gitserver shards on our Cloud
	// environment, which holds a large proportion of GitHub repositories.
	//
	// This operation returns partial results in the case of a malformed or missing
	// repository or a bad commit reference, but does not attempt to return partial
	// results when an entire shard is down. Any of these operations failing will
	// cause an error to be returned from the entire BatchLog function.

	sem := semaphore.NewWeighted(int64(32))
	g, ctx := errgroup.WithContext(ctx)

	for addr, repoCommits := range batches {
		// avoid capturing loop variable below
		addr, repoCommits := addr, repoCommits

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		g.Go(func() (err error) {
			defer sem.Release(1)

			return performLogRequestToShard(ctx, addr, repoCommits)
		})
	}

	return g.Wait()
}

func repoNamesFromRepoCommits(repoCommits []api.RepoCommit) []string {
	repoNames := make([]string, 0, len(repoCommits))
	repoNameSet := make(map[api.RepoName]struct{}, len(repoCommits))

	for _, rc := range repoCommits {
		if _, ok := repoNameSet[rc.Repo]; ok {
			continue
		}

		repoNameSet[rc.Repo] = struct{}{}
		repoNames = append(repoNames, string(rc.Repo))
	}

	return repoNames
}

func (c *clientImplementor) gitCommand(repo api.RepoName, arg ...string) GitCommand {
	if ClientMocks.LocalGitserver {
		cmd := NewLocalGitCommand(repo, arg...)
		if ClientMocks.LocalGitCommandReposDir != "" {
			cmd.ReposDir = ClientMocks.LocalGitCommandReposDir
		}
		return cmd
	}
	return &RemoteGitCommand{
		repo:   repo,
		execFn: c.httpPost,
		args:   append([]string{git}, arg...),
	}
}

func (c *clientImplementor) RequestRepoUpdate(ctx context.Context, repo api.RepoName, since time.Duration) (*protocol.RepoUpdateResponse, error) {
	req := &protocol.RepoUpdateRequest{
		Repo:  repo,
		Since: since,
	}
	resp, err := c.httpPost(ctx, repo, "repo-update", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "RepoInfo",
			Err: errors.Errorf("RepoInfo: http status %d: %s", resp.StatusCode, readResponseBody(io.LimitReader(resp.Body, 200))),
		}
	}

	var info *protocol.RepoUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

func (c *clientImplementor) RequestRepoMigrate(ctx context.Context, repo api.RepoName, from, to string) (*protocol.RepoUpdateResponse, error) {
	// We do not need to set a value for the attribute "Since" because the repo is not expected to
	// be cloned at the new gitserver instance. And for not cloned repos, this attribute is already
	// ignored.
	req := &protocol.RepoUpdateRequest{
		Repo:           repo,
		CloneFromShard: "http://" + from,
	}

	// We set "uri" to the HTTP URL of the gitserver instance that should be the new owner of this
	// "repo" based on the rendezvous hashing scheme. This way, when the gitserver instance receives
	// the request at /repo-update, it will treat it as a new clone operation and attempt to clone
	// the repo from the URL set in CloneFromShard - the gitserver instance that owns this repo based
	// on the existing hashing scheme.
	uri := "http://" + to + "/repo-update"
	resp, err := c.httpPostWithURI(ctx, repo, uri, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "RepoMigrate",
			Err: errors.Errorf("RepoMigrate: http status %d: %s", resp.StatusCode, readResponseBody(io.LimitReader(resp.Body, 200))),
		}
	}

	var info *protocol.RepoUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&info)

	return info, err
}

// MockIsRepoCloneable mocks (*Client).IsRepoCloneable for tests.
var MockIsRepoCloneable func(api.RepoName) error

func (c *clientImplementor) IsRepoCloneable(ctx context.Context, repo api.RepoName) error {
	if MockIsRepoCloneable != nil {
		return MockIsRepoCloneable(repo)
	}

	req := &protocol.IsRepoCloneableRequest{
		Repo: repo,
	}
	r, err := c.httpPost(ctx, repo, "is-repo-cloneable", req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return errors.Errorf("gitserver error (status code %d): %s", r.StatusCode, readResponseBody(r.Body))
	}

	var resp protocol.IsRepoCloneableResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Cloneable {
		return nil
	}

	// Treat all 4xx errors as not found, since we have more relaxed
	// requirements on what a valid URL is we should treat bad requests,
	// etc as not found.
	notFound := strings.Contains(resp.Reason, "not found") || strings.Contains(resp.Reason, "The requested URL returned error: 4")
	return &RepoNotCloneableErr{repo: repo, reason: resp.Reason, notFound: notFound}
}

// RepoNotCloneableErr is the error that happens when a repository can not be cloned.
type RepoNotCloneableErr struct {
	repo     api.RepoName
	reason   string
	notFound bool
}

// NotFound returns true if the repo could not be cloned because it wasn't found.
// This may be because the repo doesn't exist, or because the repo is private and
// there are insufficient permissions.
func (e *RepoNotCloneableErr) NotFound() bool {
	return e.notFound
}

func (e *RepoNotCloneableErr) Error() string {
	return fmt.Sprintf("repo not found (name=%s notfound=%v) because %s", e.repo, e.notFound, e.reason)
}

func (c *clientImplementor) RepoCloneProgress(ctx context.Context, repos ...api.RepoName) (*protocol.RepoCloneProgressResponse, error) {
	numPossibleShards := len(c.Addrs())
	shards := make(map[string]*protocol.RepoCloneProgressRequest, (len(repos)/numPossibleShards)*2) // 2x because it may not be a perfect division

	for _, r := range repos {
		addr, err := c.AddrForRepo(ctx, r)
		if err != nil {
			return nil, err
		}
		shard := shards[addr]

		if shard == nil {
			shard = new(protocol.RepoCloneProgressRequest)
			shards[addr] = shard
		}

		shard.Repos = append(shard.Repos, r)
	}

	type op struct {
		req *protocol.RepoCloneProgressRequest
		res *protocol.RepoCloneProgressResponse
		err error
	}

	ch := make(chan op, len(shards))
	for _, req := range shards {
		go func(o op) {
			var resp *http.Response
			resp, o.err = c.httpPost(ctx, o.req.Repos[0], "repo-clone-progress", o.req)
			if o.err != nil {
				ch <- o
				return
			}

			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				o.err = &url.Error{
					URL: resp.Request.URL.String(),
					Op:  "RepoCloneProgress",
					Err: errors.Errorf("RepoCloneProgress: http status %d", resp.StatusCode),
				}
				ch <- o
				return // we never get an error status code AND result
			}

			o.res = new(protocol.RepoCloneProgressResponse)
			o.err = json.NewDecoder(resp.Body).Decode(o.res)
			ch <- o
		}(op{req: req})
	}

	var err error
	res := protocol.RepoCloneProgressResponse{
		Results: make(map[api.RepoName]*protocol.RepoCloneProgress),
	}

	for i := 0; i < cap(ch); i++ {
		o := <-ch

		if o.err != nil {
			err = errors.Append(err, o.err)
			continue
		}

		for repo, info := range o.res.Results {
			res.Results[repo] = info
		}
	}

	return &res, err
}

func (c *clientImplementor) ReposStats(ctx context.Context) (map[string]*protocol.ReposStats, error) {
	stats := map[string]*protocol.ReposStats{}
	var allErr error
	for _, addr := range c.Addrs() {
		stat, err := c.doReposStats(ctx, addr)
		if err != nil {
			allErr = errors.Append(allErr, err)
		} else {
			stats[addr] = stat
		}
	}
	return stats, allErr
}

func (c *clientImplementor) doReposStats(ctx context.Context, addr string) (*protocol.ReposStats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr+"/repos-stats", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats protocol.ReposStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *clientImplementor) Remove(ctx context.Context, repo api.RepoName) error {
	// In case the repo has already been deleted from the database we need to pass
	// the old name in order to land on the correct gitserver instance.
	addr, err := c.AddrForRepo(ctx, api.UndeletedRepoName(repo))
	if err != nil {
		return err
	}
	return c.RemoveFrom(ctx, repo, addr)
}

func (c *clientImplementor) RemoveFrom(ctx context.Context, repo api.RepoName, from string) error {
	b, err := json.Marshal(&protocol.RepoDeleteRequest{
		Repo: repo,
	})
	if err != nil {
		return err
	}

	uri := "http://" + from + "/delete"
	resp, err := c.do(ctx, repo, "POST", uri, b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "RepoRemove",
			Err: errors.Errorf("RepoRemove: http status %d: %s", resp.StatusCode, readResponseBody(io.LimitReader(resp.Body, 200))),
		}
	}
	return nil
}

// httpPost will apply the MD5 hashing scheme on the repo name to determine the gitserver instance
// to which the HTTP POST request is sent. To use the rendezvous hashing scheme, see
// httpPostWithURI.
func (c *clientImplementor) httpPost(ctx context.Context, repo api.RepoName, op string, payload any) (resp *http.Response, err error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	addrForRepo, err := c.AddrForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	uri := "http://" + addrForRepo + "/" + op
	return c.do(ctx, repo, "POST", uri, b)
}

// httpPostWithURI does not apply any transformations to the given URI. This allows the consumer to
// use the predetermined hashing scheme (md5 or rendezvous) of their choice to derive the gitserver
// instance to which the HTTP POST request is sent.
func (c *clientImplementor) httpPostWithURI(ctx context.Context, repo api.RepoName, uri string, payload any) (resp *http.Response, err error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, repo, "POST", uri, b)
}

// do performs a request to a gitserver instance based on the address in the uri argument.
//
//nolint:unparam // unparam complains that `method` always has same value across call-sites, but that's OK
func (c *clientImplementor) do(ctx context.Context, repo api.RepoName, method, uri string, payload []byte) (resp *http.Response, err error) {
	parsedURL, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, errors.Wrap(err, "do")
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Client.do")
	defer func() {
		span.LogKV("repo", string(repo), "method", method, "path", parsedURL.Path)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	req, err := http.NewRequest(method, uri, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req = req.WithContext(ctx)

	if c.HTTPLimiter != nil {
		c.HTTPLimiter.Acquire()
		defer c.HTTPLimiter.Release()
		span.LogKV("event", "Acquired HTTP limiter")
	}

	req, ht := nethttp.TraceRequest(span.Tracer(), req,
		nethttp.OperationName("Gitserver Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	return c.httpClient.Do(req)
}

func (c *clientImplementor) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	resp, err := c.httpPost(ctx, req.Repo, "create-commit-from-patch", req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("reading gitserver create-commit-from-patch response", sglog.Error(err))
		return "", &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "CreateCommitFromPatch",
			Err: errors.Errorf("CreateCommitFromPatch: http status %d, %s", resp.Status, readResponseBody(resp.Body)),
		}
	}

	var res protocol.CreateCommitFromPatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		c.logger.Warn("decoding gitserver create-commit-from-patch response", sglog.Error(err))
		return "", &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "CreateCommitFromPatch",
			Err: errors.Errorf("CreateCommitFromPatch: http status %d, %v", resp.StatusCode, err),
		}
	}

	if res.Error != nil {
		return res.Rev, res.Error
	}
	return res.Rev, nil
}

func (c *clientImplementor) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error) {
	if ClientMocks.GetObject != nil {
		return ClientMocks.GetObject(repo, objectName)
	}

	req := protocol.GetObjectRequest{
		Repo:       repo,
		ObjectName: objectName,
	}
	resp, err := c.httpPost(ctx, req.Repo, "commands/get-object", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("reading gitserver get-object response", sglog.Error(err))
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "GetObject",
			Err: errors.Errorf("GetObject: http status %d, %s", resp.StatusCode, readResponseBody(resp.Body)),
		}
	}

	var res protocol.GetObjectResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		c.logger.Warn("decoding gitserver get-object response", sglog.Error(err))
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "GetObject",
			Err: errors.Errorf("GetObject: http status %d, failed to decode response body: %v", resp.StatusCode, err),
		}
	}

	return &res.Object, nil
}

var ambiguousArgPattern = lazyregexp.New(`ambiguous argument '([^']+)'`)

func (c *clientImplementor) ResolveRevisions(ctx context.Context, repo api.RepoName, revs []protocol.RevisionSpecifier) ([]string, error) {
	args := append([]string{"rev-parse"}, revsToGitArgs(revs)...)

	cmd := c.gitCommand(repo, args...)
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, err
		}
		if match := ambiguousArgPattern.FindSubmatch(stderr); match != nil {
			return nil, &gitdomain.RevisionNotFoundError{Repo: repo, Spec: string(match[1])}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (stderr: %q)", cmd.Args(), stderr))
	}

	return strings.Fields(string(stdout)), nil
}

func revsToGitArgs(revSpecs []protocol.RevisionSpecifier) []string {
	args := make([]string, 0, len(revSpecs))
	for _, r := range revSpecs {
		if r.RevSpec != "" {
			args = append(args, r.RevSpec)
		} else if r.RefGlob != "" {
			args = append(args, "--glob="+r.RefGlob)
		} else if r.ExcludeRefGlob != "" {
			args = append(args, "--exclude="+r.ExcludeRefGlob)
		} else {
			args = append(args, "HEAD")
		}
	}

	// If revSpecs is empty, git treats it as equivalent to HEAD
	if len(revSpecs) == 0 {
		args = append(args, "HEAD")
	}
	return args
}

// shouldUseRendezvousHashing returns true if rendezvous hashing is to be used to find
// an address of gitserver instance for a given repo
func shouldUseRendezvousHashing(ctx context.Context, db database.DB, repo string) (bool, error) {
	cursor, err := migration.GetCursor(ctx, db)
	if err != nil {
		return false, err
	}
	if cursor == "" {
		return false, nil
	}

	// Migration is in progress or finished, if the name is less than or equal to cursor -- use rendezvous
	return repo <= cursor, nil
}

// getPinnedRepoAddr returns true and gitserver address if given repo is pinned.
// Otherwise, if repo is not pinned -- false and empty string are returned
func getPinnedRepoAddr(repo string, pinnedServers map[string]string) (bool, string) {
	pinned, found := pinnedServers[repo]
	return found, pinned
}

// readResponseBody will attempt to read the body of the HTTP response and return it as a
// string. However, in the unlikely scenario that it fails to read the body, it will encode and
// return the error message as a string.
//
// This allows us to use this function directly without yet another if err != nil check. As a
// result, this function should **only** be used when we're attempting to return the body's content
// as part of an error. In such scenarios we don't need to return the potential error from reading
// the body, but can get away with returning that error as a string itself.
//
// This is an unusual pattern of not returning an error. Be careful of replicating this in other
// parts of the code.
func readResponseBody(body io.Reader) string {
	content, err := io.ReadAll(body)
	if err != nil {
		return fmt.Sprintf("failed to read response body, error: %v", err)
	}

	// strings.TrimSpace is needed to remove trailing \n characters that is added by the
	// server. We use http.Error in the server which in turn uses fmt.Fprintln to format
	// the error message. And in translation that newline gets escapted into a \n
	// character.  For what the error message would look in the UI without
	// strings.TrimSpace, see attached screenshots in this pull request:
	// https://github.com/sourcegraph/sourcegraph/pull/39358.
	return strings.TrimSpace(string(content))
}

func pinnedReposFromConfig() map[string]string {
	cfg := conf.Get()
	if cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.GitServerPinnedRepos != nil {
		return cfg.ExperimentalFeatures.GitServerPinnedRepos
	}
	return map[string]string{}
}
