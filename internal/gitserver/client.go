package gitserver

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
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
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/inconshreveable/log15"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/go-rendezvous"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
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
	RepoInfo                func(ctx context.Context, repos ...api.RepoName) (*protocol.RepoInfoResponse, error)
	Archive                 func(ctx context.Context, repo api.RepoName, opt ArchiveOptions) (_ io.ReadCloser, err error)
	LocalGitserver          bool
	LocalGitCommandReposDir string
}

// AddrsMock is a mock for Addrs() function. It is separated from ClientMocks
// because it is not intended to be cleared when other mocks should be.
// This mock should be initialized during tests initialization so that
// gitserver client always contain address of a local machine during tests
// and tests which use gitserver client can pass successfully
var AddrsMock func() []string

// ResetClientMocks clears the mock functions set on Mocks (so that subsequent
// tests don't inadvertently use them).
func ResetClientMocks() {
	ClientMocks = emptyClientMocks
}

// NewClient returns a new gitserver.Client instantiated with default arguments
// and httpcli.Doer.
func NewClient(db database.DB) *ClientImplementor {
	return &ClientImplementor{
		addrs: func() []string {
			return conf.Get().ServiceConnections().GitServers
		},
		pinned: func() map[string]string {
			cfg := conf.Get()
			if cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.GitServerPinnedRepos != nil {
				return cfg.ExperimentalFeatures.GitServerPinnedRepos
			}
			return map[string]string{}
		},
		db:          db,
		HTTPClient:  defaultDoer,
		HTTPLimiter: defaultLimiter,
		// Use the binary name for UserAgent. This should effectively identify
		// which service is making the request (excluding requests proxied via the
		// frontend internal API)
		UserAgent:  filepath.Base(os.Args[0]),
		operations: getOperations(),
	}
}

func NewTestClient(cli httpcli.Doer, db database.DB, addrs []string) *ClientImplementor {
	return &ClientImplementor{
		addrs: func() []string {
			return addrs
		},
		pinned: func() map[string]string {
			// nothing needs to be pinned for the tests
			return conf.Get().ExperimentalFeatures.GitServerPinnedRepos
		},
		HTTPClient:  cli,
		HTTPLimiter: parallel.NewRun(500),
		// Use the binary name for UserAgent. This should effectively identify
		// which service is making the request (excluding requests proxied via the
		// frontend internal API)
		UserAgent:  filepath.Base(os.Args[0]),
		db:         db,
		operations: newOperations(&observation.TestContext),
	}
}

// ClientImplementor is a gitserver client.
type ClientImplementor struct {
	// HTTP client to use
	HTTPClient httpcli.Doer

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run

	// addrs is a function which should return the addresses for gitservers. It
	// is called each time a request is made. The function must be safe for
	// concurrent use. It may return different results at different times.
	addrs func() []string

	// pinned holds a map of repositories(key) pinned to a particular gitserver instance(value). This function
	// should query the conf to fetch a fresh map of pinned repos, so that we don't have to proactively watch for conf changes
	// and sync the pinned map.
	pinned func() map[string]string

	// UserAgent is a string identifying who the client is. It will be logged in
	// the telemetry in gitserver.
	UserAgent string

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

//go:generate ../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/gitserver -i Client -o mock_client.go
type Client interface {
	// AddrForRepo returns the gitserver address to use for the given repo name.
	AddrForRepo(context.Context, api.RepoName) (string, error)

	// Addrs returns the addresses for gitservers. It is safe for concurrent
	// use. It may return different results at different times.
	Addrs() []string

	// Archive produces an archive from a Git repository.
	Archive(context.Context, api.RepoName, ArchiveOptions) (io.ReadCloser, error)

	// ArchiveURL returns a URL from which an archive of the given Git repository can
	// be downloaded from.
	ArchiveURL(context.Context, api.RepoName, ArchiveOptions) (*url.URL, error)

	// BatchLog invokes the given callback with the `git log` output for a batch of repository
	// and commit pairs. If the invoked callback returns a non-nil error, the operation will begin
	// to abort processing further results.
	BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) error

	// BlameFile returns Git blame information about a file.
	BlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions, checker authz.SubRepoPermissionChecker) ([]*Hunk, error)

	// GitCommand creates a new GitCommand.
	GitCommand(repo api.RepoName, args ...string) GitCommand

	// CreateCommitFromPatch will attempt to create a commit from a patch
	// If possible, the error returned will be of type protocol.CreateCommitFromPatchError
	CreateCommitFromPatch(context.Context, protocol.CreateCommitFromPatchRequest) (string, error)

	// GetObject fetches git object data in the supplied repo
	GetObject(_ context.Context, _ api.RepoName, objectName string) (*gitdomain.GitObject, error)

	// IsRepoCloneable returns nil if the repository is cloneable.
	IsRepoCloneable(context.Context, api.RepoName) error

	IsRepoCloned(context.Context, api.RepoName) (bool, error)

	// ListCloned lists all cloned repositories
	ListCloned(context.Context) ([]string, error)

	// ListGitolite lists Gitolite repositories.
	ListGitolite(_ context.Context, gitoliteHost string) ([]*gitolite.Repo, error)

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

	// RepoInfo retrieves information about one or more repositories on gitserver.
	//
	// The repository not existing is not an error; in that case, RepoInfoResponse.Results[i].Cloned
	// will be false and the error will be nil.
	//
	// If multiple errors occurred, an incomplete result is returned along with an
	// error.errors.
	RepoInfo(context.Context, ...api.RepoName) (*protocol.RepoInfoResponse, error)

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

	// DiffPath returns a position-ordered slice of changes (additions or deletions)
	// of the given path between the given source and target commits.
	DiffPath(ctx context.Context, repo api.RepoName, sourceCommit, targetCommit, path string, checker authz.SubRepoPermissionChecker) ([]*diff.Hunk, error)

	// ReadDir reads the contents of the named directory at commit.
	ReadDir(ctx context.Context, db database.DB, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error)
}

func (c *ClientImplementor) Addrs() []string {
	if AddrsMock != nil {
		return AddrsMock()
	}
	return c.addrs()
}

func (c *ClientImplementor) AddrForRepo(ctx context.Context, repo api.RepoName) (string, error) {
	addrs := c.Addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	return AddrForRepo(ctx, c.UserAgent, c.db, repo, GitServerAddresses{
		Addresses:     addrs,
		PinnedServers: c.pinned(),
	})
}

func (c *ClientImplementor) RendezvousAddrForRepo(repo api.RepoName) string {
	addrs := c.Addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	if repoPinned, addr := getPinnedRepoAddr(string(repo), c.pinned()); repoPinned {
		return addr
	}
	return RendezvousAddrForRepo(repo, addrs)
}

// addrForKey returns the gitserver address to use for the given string key,
// which is hashed for sharding purposes.
func (c *ClientImplementor) addrForKey(key string) string {
	addrs := c.Addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	return addrForKey(key, addrs)
}

var addrForRepoInvoked = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_invoked",
	Help: "Number of times gitserver.AddrForRepo was invoked",
}, []string{"user_agent"})

// AddrForRepoCounter is used to track the number of times AddrForRepo is called
// and is used to determine if we can read the gitserver location from the database.
// See AddrForRepo for more details.
var AddrForRepoCounter uint64

// addrForRepoAddrMismatch is used to count the number of times the state of
// the gitserver_repos table and the hashing algorithm disagree.
var addrForRepoAddrMismatch = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_addr_mismatch",
	Help: "Number of times the gitserver_repos state and the result of gitserver.AddrForRepo are mismatched",
}, []string{"user_agent"})

// addrForRepoAddrFromDB is used to count the number of times we called the
// database to get the gitserver address for a repo.
var addrForRepoAddrFromDB = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_gitserver_addr_for_repo_addr_from_db",
	Help: "Number of times the gitserver address for a repo is retrieved from the database",
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

	var shardID string
	useRendezvous, err := shouldUseRendezvousHashing(ctx, db, rs)
	if err != nil {
		return "", err
	}
	if useRendezvous {
		shardID, err = RendezvousAddrForRepo(repo, addresses.Addresses), nil
	} else {
		shardID, err = addrForKey(rs, addresses.Addresses), nil
	}
	if err != nil {
		return "", err
	}

	// This is an experiment to determine the impact on using the database to determine the location of a repo.
	// It is meant to be used exclusively on Cloud and the rate will be increased progressively.
	// Once we determine the impact of this experiment, we can remove it.
	cfg := conf.Get()
	if envvar.SourcegraphDotComMode() && cfg.ExperimentalFeatures != nil && cfg.ExperimentalFeatures.EnableGitserverClientLookupTable {
		// get the rate from the configuration. The rate is a percentage, and defaults to 0.
		var rate uint64 = uint64(conf.Get().ExperimentalFeatures.GitserverClientLookupTableRate)
		if rate > 100 {
			rate = 0
		}

		// We are using a modulo operation to spread the calls to the database across the rate.
		var mod uint64
		if rate > 0 {
			mod = 100 / rate
		}

		if rate != 0 && atomic.AddUint64(&AddrForRepoCounter, 1)%mod == 0 {
			span, ctx := ot.StartSpanFromContext(ctx, "GitserverClient.AddrForRepoFromDB")
			span.SetTag("Repo", repo)
			defer func() {
				if err != nil {
					ext.Error.Set(span, true)
					span.LogFields(otlog.Error(err))
				}
				span.Finish()
			}()

			addrForRepoAddrFromDB.WithLabelValues(userAgent).Inc()

			gr, err := db.GitserverRepos().GetByName(ctx, repo)
			switch {
			case err == nil:
				// if there is a difference between the database state and the hashing result
				// we should observe it.
				if gr.ShardID != shardID {
					addrForRepoAddrMismatch.WithLabelValues(userAgent).Inc()
				}
			case !errors.Is(err, sql.ErrNoRows):
				log15.Warn("gitserver.AddrForRepo: failed to get gitserver repo from the database", "repo", repo, "err", err)
			}
		}
	}

	return shardID, nil
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
	Treeish   string     // the tree or commit to produce an archive for
	Format    string     // format of the resulting archive (usually "tar" or "zip")
	Pathspecs []Pathspec // if nonempty, only include these pathspecs.
}

type BatchLogOptions protocol.BatchLogRequest

func (opts BatchLogOptions) LogFields() []log.Field {
	return []log.Field{
		log.Int("numRepoCommits", len(opts.RepoCommits)),
		log.String("Format", opts.Format),
	}
}

// Pathspec is a git term for a pattern that matches paths using glob-like syntax.
// https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-aiddefpathspecapathspec
type Pathspec string

// PathspecLiteral constructs a pathspec that matches a path without interpreting "*" or "?" as special
// characters.
func PathspecLiteral(s string) Pathspec { return Pathspec(":(literal)" + s) }

// PathspecSuffix constructs a pathspec that matches paths ending with the given suffix (useful for
// matching paths by basename).
func PathspecSuffix(s string) Pathspec { return Pathspec("*" + s) }

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

// ArchiveURL returns a URL from which an archive of the given Git repository can
// be downloaded from.
func (c *ClientImplementor) ArchiveURL(ctx context.Context, repo api.RepoName, opt ArchiveOptions) (*url.URL, error) {
	q := url.Values{
		"repo":    {string(repo)},
		"treeish": {opt.Treeish},
		"format":  {opt.Format},
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

func (c *ClientImplementor) Archive(ctx context.Context, repo api.RepoName, opt ArchiveOptions) (_ io.ReadCloser, err error) {
	if ClientMocks.Archive != nil {
		return ClientMocks.Archive(ctx, repo, opt)
	}
	span, ctx := ot.StartSpanFromContext(ctx, "Git: Archive")
	span.SetTag("Repo", repo)
	span.SetTag("Treeish", opt.Treeish)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	// Check that ctx is not expired.
	if err := ctx.Err(); err != nil {
		deadlineExceededCounter.Inc()
		return nil, err
	}

	u, err := c.ArchiveURL(ctx, repo, opt)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, repo, "POST", u.String(), nil)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return &archiveReader{
			base: &cmdReader{
				rc:      resp.Body,
				trailer: resp.Trailer,
			},
			repo: repo,
			spec: opt.Treeish,
		}, nil
	case http.StatusNotFound:
		var payload protocol.NotFoundPayload
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
		return nil, &badRequestError{
			error: &gitdomain.RepoNotExistError{
				Repo:            repo,
				CloneInProgress: payload.CloneInProgress,
				CloneProgress:   payload.CloneProgress,
			},
		}
	default:
		resp.Body.Close()
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
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

func (c *ClientImplementor) Search(ctx context.Context, args *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, err error) {
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

func (c *ClientImplementor) P4Exec(ctx context.Context, host, user, password string, args ...string) (_ io.ReadCloser, _ http.Header, errRes error) {
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

	switch resp.StatusCode {
	case http.StatusOK:
		return resp.Body, resp.Trailer, nil

	default:
		// Read response body at best effort
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, errors.Errorf("unexpected status code: %d - %s", resp.StatusCode, body)
	}
}

var deadlineExceededCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_gitserver_client_deadline_exceeded",
	Help: "Times that Client.sendExec() returned context.DeadlineExceeded",
})

// BatchLog invokes the given callback with the `git log` output for a batch of repository
// and commit pairs. If the invoked callback returns a non-nil error, the operation will begin
// to abort processing further results.
func (c *ClientImplementor) BatchLog(ctx context.Context, opts BatchLogOptions, callback BatchLogCallback) (err error) {
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

		// TODO(efritz) - remove after 3.39 branch cut
		if resp.StatusCode == http.StatusNotFound {
			// Frontend and gitserver may be rolling out. Fall back to issuing one
			// command per item in the batch via the original /exec endpoint. We
			// inline the same behavior as BatchLog here as this is throw-away code.

			for _, repoCommit := range repoCommits {
				content, err := func() (string, error) {
					reader, err := c.execReader(ctx, repoCommit.Repo, []string{"log", "-n", "1", "--name-only", opts.Format, string(repoCommit.CommitID)})
					if err != nil {
						return "", errors.Wrap(err, "execReader")
					}

					content, err := io.ReadAll(reader)
					if err != nil {
						return "", errors.Wrap(err, "io.ReadAll")
					}

					return string(content), nil
				}()

				rawResult := RawBatchLogResult{
					Stdout: content,
					Error:  err,
				}
				if err := callback(repoCommit, rawResult); err != nil {
					return errors.Wrap(err, "commitLogCallback")
				}
			}

			return nil
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
			return errors.Newf("http status %d: %s", resp.StatusCode, body)
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

func (c *ClientImplementor) GitCommand(repo api.RepoName, arg ...string) GitCommand {
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

func (c *ClientImplementor) ListGitolite(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	// The gitserver calls the shared Gitolite server in response to this request, so
	// we need to only call a single gitserver (or else we'd get duplicate results).
	addr := c.addrForKey(gitoliteHost)
	req, err := http.NewRequest("GET", "http://"+addr+"/list-gitolite?gitolite="+url.QueryEscape(gitoliteHost), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&list)
	return list, err
}

func (c *ClientImplementor) ListCloned(ctx context.Context) ([]string, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		err   error
		repos []string
	)
	addrs := c.Addrs()
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			r, e := c.doListOne(ctx, "?cloned", addr)

			// Only include repos that belong on addr.
			if len(r) > 0 {
				filtered := r[:0]
				for _, repo := range r {
					if addrForKey(repo, addrs) == addr {
						filtered = append(filtered, repo)
					}
				}
				r = filtered
			}
			mu.Lock()
			if e != nil {
				err = e
			}
			repos = append(repos, r...)
			mu.Unlock()
		}(addr)
	}
	wg.Wait()
	return repos, err
}

func (c *ClientImplementor) doListOne(ctx context.Context, urlSuffix, addr string) ([]string, error) {
	req, err := http.NewRequest("GET", "http://"+addr+"/list"+urlSuffix, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list []string
	err = json.NewDecoder(resp.Body).Decode(&list)
	return list, err
}

func (c *ClientImplementor) RequestRepoUpdate(ctx context.Context, repo api.RepoName, since time.Duration) (*protocol.RepoUpdateResponse, error) {
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, &url.Error{URL: resp.Request.URL.String(), Op: "RepoInfo", Err: errors.Errorf("RepoInfo: http status %d: %s", resp.StatusCode, body)}
	}

	var info *protocol.RepoUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

func (c *ClientImplementor) RequestRepoMigrate(ctx context.Context, repo api.RepoName, from, to string) (*protocol.RepoUpdateResponse, error) {
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, &url.Error{
			URL: resp.Request.URL.String(),
			Op:  "RepoMigrate",
			Err: errors.Errorf("RepoMigrate: http status %d: %s", resp.StatusCode, body),
		}
	}

	var info *protocol.RepoUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&info)

	return info, err
}

// MockIsRepoCloneable mocks (*Client).IsRepoCloneable for tests.
var MockIsRepoCloneable func(api.RepoName) error

func (c *ClientImplementor) IsRepoCloneable(ctx context.Context, repo api.RepoName) error {
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return errors.Errorf("gitserver error (status code %d): %s", r.StatusCode, string(body))
	}

	var resp protocol.IsRepoCloneableResponse
	if err := json.Unmarshal(body, &resp); err != nil {
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

func (c *ClientImplementor) IsRepoCloned(ctx context.Context, repo api.RepoName) (bool, error) {
	req := &protocol.IsRepoClonedRequest{
		Repo: repo,
	}
	resp, err := c.httpPost(ctx, repo, "is-repo-cloned", req)
	if err != nil {
		return false, err
	}
	// no need to defer, we aren't using the body.
	resp.Body.Close()
	var cloned bool
	if resp.StatusCode == http.StatusOK {
		cloned = true
	}
	return cloned, nil
}

func (c *ClientImplementor) RepoCloneProgress(ctx context.Context, repos ...api.RepoName) (*protocol.RepoCloneProgressResponse, error) {
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

func (c *ClientImplementor) RepoInfo(ctx context.Context, repos ...api.RepoName) (*protocol.RepoInfoResponse, error) {
	if ClientMocks.RepoInfo != nil {
		return ClientMocks.RepoInfo(ctx, repos...)
	}

	numPossibleShards := len(c.Addrs())
	shards := make(map[string]*protocol.RepoInfoRequest, (len(repos)/numPossibleShards)*2) // 2x because it may not be a perfect division

	for _, r := range repos {
		addr, err := c.AddrForRepo(ctx, r)
		if err != nil {
			return nil, err
		}
		shard := shards[addr]

		if shard == nil {
			shard = new(protocol.RepoInfoRequest)
			shards[addr] = shard
		}

		shard.Repos = append(shard.Repos, r)
	}

	type op struct {
		req *protocol.RepoInfoRequest
		res *protocol.RepoInfoResponse
		err error
	}

	ch := make(chan op, len(shards))
	for _, req := range shards {
		go func(o op) {
			var resp *http.Response
			resp, o.err = c.httpPost(ctx, o.req.Repos[0], "repos", o.req)
			if o.err != nil {
				ch <- o
				return
			}

			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				o.err = &url.Error{
					URL: resp.Request.URL.String(),
					Op:  "RepoInfo",
					Err: errors.Errorf("RepoInfo: http status %d", resp.StatusCode),
				}
				ch <- o
				return // we never get an error status code AND result
			}

			o.res = new(protocol.RepoInfoResponse)
			o.err = json.NewDecoder(resp.Body).Decode(o.res)
			ch <- o
		}(op{req: req})
	}

	var err error
	res := protocol.RepoInfoResponse{
		Results: make(map[api.RepoName]*protocol.RepoInfo),
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

func (c *ClientImplementor) ReposStats(ctx context.Context) (map[string]*protocol.ReposStats, error) {
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

func (c *ClientImplementor) doReposStats(ctx context.Context, addr string) (*protocol.ReposStats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr+"/repos-stats", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
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

func (c *ClientImplementor) Remove(ctx context.Context, repo api.RepoName) error {
	// In case the repo has already been deleted from the database we need to pass
	// the old name in order to land on the correct gitserver instance.
	addr, err := c.AddrForRepo(ctx, api.UndeletedRepoName(repo))
	if err != nil {
		return err
	}
	return c.RemoveFrom(ctx, repo, addr)
}

func (c *ClientImplementor) RemoveFrom(ctx context.Context, repo api.RepoName, from string) error {
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
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return &url.Error{URL: resp.Request.URL.String(), Op: "RepoRemove", Err: errors.Errorf("RepoRemove: http status %d: %s", resp.StatusCode, string(body))}
	}
	return nil
}

// httpPost will apply the MD5 hashing scheme on the repo name to determine the gitserver instance
// to which the HTTP POST request is sent. To use the rendezvous hashing scheme, see
// httpPostWithURI.
func (c *ClientImplementor) httpPost(ctx context.Context, repo api.RepoName, op string, payload any) (resp *http.Response, err error) {
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
func (c *ClientImplementor) httpPostWithURI(ctx context.Context, repo api.RepoName, uri string, payload any) (resp *http.Response, err error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, repo, "POST", uri, b)
}

//nolint:unparam // unparam complains that `method` always has same value across call-sites, but that's OK
// do performs a request to a gitserver instance based on the address in the uri argument.
func (c *ClientImplementor) do(ctx context.Context, repo api.RepoName, method, uri string, payload []byte) (resp *http.Response, err error) {
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
	req.Header.Set("User-Agent", c.UserAgent)
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

	return c.HTTPClient.Do(req)
}

func (c *ClientImplementor) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	resp, err := c.httpPost(ctx, req.Repo, "create-commit-from-patch", req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log15.Warn("reading gitserver create-commit-from-patch response", "err", err.Error())
		return "", &url.Error{URL: resp.Request.URL.String(), Op: "CreateCommitFromPatch", Err: errors.Errorf("CreateCommitFromPatch: http status %d %s", resp.StatusCode, err.Error())}
	}

	var res protocol.CreateCommitFromPatchResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		log15.Warn("decoding gitserver create-commit-from-patch response", "err", err.Error())
		return "", &url.Error{URL: resp.Request.URL.String(), Op: "CreateCommitFromPatch", Err: errors.Errorf("CreateCommitFromPatch: http status %d %s", resp.StatusCode, string(data))}
	}

	if res.Error != nil {
		return res.Rev, res.Error
	}
	return res.Rev, nil
}

func (c *ClientImplementor) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error) {
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log15.Warn("reading gitserver get-object response", "err", err.Error())
		return nil, &url.Error{URL: resp.Request.URL.String(), Op: "GetObject", Err: errors.Errorf("GetObject: http status %d %s", resp.StatusCode, err.Error())}
	}

	var res protocol.GetObjectResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		log15.Warn("decoding gitserver get-object response", "err", err.Error())
		return nil, &url.Error{URL: resp.Request.URL.String(), Op: "GetObject", Err: errors.Errorf("GetObject: http status %d %s", resp.StatusCode, string(data))}
	}

	return &res.Object, nil
}

var ambiguousArgPattern = lazyregexp.New(`ambiguous argument '([^']+)'`)

func (c *ClientImplementor) ResolveRevisions(ctx context.Context, repo api.RepoName, revs []protocol.RevisionSpecifier) ([]string, error) {
	args := append([]string{"rev-parse"}, revsToGitArgs(revs)...)

	cmd := c.GitCommand(repo, args...)
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
