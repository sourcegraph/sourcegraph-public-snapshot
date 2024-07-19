package gitserver

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/go-diff/diff"
	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/connection"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ Client = &clientImplementor{}

// ClientSource is a source of gitserver.Client instances.
// It allows for mocking out the client source in tests.
type ClientSource interface {
	// ClientForRepo returns a Client for the given repo.
	ClientForRepo(ctx context.Context, repo api.RepoName) (proto.GitserverServiceClient, error)
	// AddrForRepo returns the address of the gitserver for the given repo.
	AddrForRepo(ctx context.Context, repo api.RepoName) string
	// Address the current list of gitserver addresses.
	Addresses() []AddressWithClient
	// GetAddressWithClient returns the address and client for a gitserver instance.
	// It returns nil if there's no server with that address
	GetAddressWithClient(addr string) AddressWithClient
}

type clientSource struct{}

func (cs *clientSource) ClientForRepo(ctx context.Context, repo api.RepoName) (proto.GitserverServiceClient, error) {
	conn, err := connection.GlobalConns.ConnForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	return clientForConn(conn), nil
}

func (cs *clientSource) AddrForRepo(ctx context.Context, repo api.RepoName) string {
	return connection.GlobalConns.AddrForRepo(ctx, repo)
}

func (cs *clientSource) Addresses() []AddressWithClient {
	conns := connection.GlobalConns.Addresses()
	addrs := make([]AddressWithClient, len(conns))
	for i, addr := range conns {
		conn, err := addr.GRPCConn()
		addrs[i] = &connAndErr{
			address: addr.Address(),
			conn:    conn,
			err:     err,
		}
	}
	return addrs
}

func (cs *clientSource) GetAddressWithClient(addr string) AddressWithClient {
	ac := connection.GlobalConns.GetAddressWithConn(addr)

	conn, err := ac.GRPCConn()
	return &connAndErr{
		address: ac.Address(),
		conn:    conn,
		err:     err,
	}
}

// NewClient returns a new gitserver.Client.
// See Client.Scoped() for info on scoped clients.
func NewClient(scope string) Client {
	logger := sglog.Scoped("GitserverClient")
	return &clientImplementor{
		logger:              logger,
		scope:               scope,
		operations:          getOperations(),
		clientSource:        &clientSource{},
		subRepoPermsChecker: authz.DefaultSubRepoPermsChecker,
	}
}

// NewTestClient returns a test client that will us
func NewTestClient(t testing.TB) TestClient {
	logger := logtest.Scoped(t)

	return &clientImplementor{
		logger:              logger,
		scope:               fmt.Sprintf("gitserver.test.%s", t.Name()),
		operations:          newOperations(observation.ContextWithLogger(logger, observation.TestContextTB(t))),
		clientSource:        NewTestClientSource(t, nil),
		subRepoPermsChecker: authz.DefaultSubRepoPermsChecker,
	}
}

type TestClient interface {
	Client
	WithChecker(authz.SubRepoPermissionChecker) TestClient
	WithClientSource(ClientSource) TestClient
}

func (c *clientImplementor) WithChecker(checker authz.SubRepoPermissionChecker) TestClient {
	c.subRepoPermsChecker = checker
	return c
}

func (c *clientImplementor) WithClientSource(cs ClientSource) TestClient {
	c.clientSource = cs
	return c
}

// NewMockClientWithExecReader return new MockClient with provided mocked
// behaviour of ExecReader function.
func NewMockClientWithExecReader(checker authz.SubRepoPermissionChecker, execReader func(context.Context, api.RepoName, []string) (io.ReadCloser, error)) *MockClient {
	client := NewMockClient()
	// NOTE: This hook is the same as DiffFunc, but with `execReader` used above
	client.DiffFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, opts DiffOptions) (*DiffFileIterator, error) {
		if opts.Base == DevNullSHA {
			opts.RangeType = ".."
		} else if opts.RangeType != ".." {
			opts.RangeType = "..."
		}

		rangeSpec := opts.Base + opts.RangeType + opts.Head
		if strings.HasPrefix(rangeSpec, "-") || strings.HasPrefix(rangeSpec, ".") {
			return nil, errors.Errorf("invalid diff range argument: %q", rangeSpec)
		}

		// Here is where all the mocking happens!
		rdr, err := execReader(ctx, repo, append([]string{
			"diff",
			"--find-renames",
			"--full-index",
			"--inter-hunk-context=3",
			"--no-prefix",
			rangeSpec,
			"--",
		}, opts.Paths...))
		if err != nil {
			return nil, errors.Wrap(err, "executing git diff")
		}

		return &DiffFileIterator{
			onClose:        func() { rdr.Close() },
			mfdr:           diff.NewMultiFileDiffReader(rdr),
			fileFilterFunc: getFilterFunc(ctx, checker, repo),
		}, nil
	})

	return client
}

// clientImplementor is a gitserver client.
type clientImplementor struct {
	// the current scope of the client.
	scope string

	// logger is used for all logging and logger creation
	logger sglog.Logger

	// operations are used for internal observability
	operations *operations

	// clientSource is used to get the corresponding gprc client or address for a given repository
	clientSource ClientSource

	// subRepoPermsChecker is sub-repository permissions checker. This will
	// usually be authz.DefaultSubRepoPermsChecker, at least until that global is removed.
	subRepoPermsChecker authz.SubRepoPermissionChecker
}

func (c *clientImplementor) Scoped(scope string) Client {
	return &clientImplementor{
		logger:       c.logger,
		scope:        appendScope(c.scope, scope),
		operations:   c.operations,
		clientSource: c.clientSource,
	}
}

func appendScope(existing, new string) string {
	if existing == "" {
		return new
	}
	return existing + "." + new
}

type HunkReader interface {
	Read() (*gitdomain.Hunk, error)
	Close() error
}

// BlameOptions configures a blame.
type BlameOptions struct {
	NewestCommit     api.CommitID `json:",omitempty" url:",omitempty"`
	IgnoreWhitespace bool         `json:",omitempty" url:",omitempty"`
	Range            *BlameRange  `json:",omitempty" url:",omitempty"`
}

func (o *BlameOptions) Attrs() []attribute.KeyValue {
	kvs := []attribute.KeyValue{
		attribute.String("newestCommit", string(o.NewestCommit)),
		attribute.Bool("ignoreWhitespace", o.IgnoreWhitespace),
	}
	if o.Range != nil {
		kvs = append(kvs, o.Range.Attrs()...)
	}
	return kvs
}

type BlameRange struct {
	StartLine int `json:",omitempty" url:",omitempty"` // 1-indexed start line (or 0 for beginning of file)
	EndLine   int `json:",omitempty" url:",omitempty"` // 1-indexed end line (or 0 for end of file)
}

func (o *BlameRange) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("startLine", o.StartLine),
		attribute.Int("endLine", o.EndLine),
	}
}

// ListRefsOpts are additional options passed to ListRefs.
type ListRefsOpts struct {
	// If true, only heads are returned. Can be combined with TagsOnly.
	HeadsOnly bool
	// If true, only tags are returned. Can be combined with HeadsOnly.
	TagsOnly bool
	// If set, only return refs that point at the given commit sha. Multiple
	// values are ORed together.
	PointsAtCommit []api.CommitID
	// If set, only return refs that contain the given commit sha in their history.
	Contains api.CommitID
}

// ArchiveOptions contains options for the Archive func.
type ArchiveOptions struct {
	// Treeish is the tree or commit to produce an archive for
	Treeish string
	// Format is the format of the resulting archive (usually "tar" or "zip")
	Format ArchiveFormat
	// Paths is a list of paths to include in the archive. If empty, the entire
	// repository is included.
	//
	// Note: The Path strings are not guaranteed to be UTF-8 encoded, as file paths
	// on Linux can be arbitrary byte sequences. Users should take care to validate / sanitize
	// paths if necessary.
	Paths []string
}

func (a *ArchiveOptions) Attrs() []attribute.KeyValue {
	pathAttrs := make([]string, len(a.Paths))
	for i, path := range a.Paths {
		pathAttrs[i] = string(path)
	}
	return []attribute.KeyValue{
		attribute.String("treeish", a.Treeish),
		attribute.String("format", string(a.Format)),
		attribute.StringSlice("paths", pathAttrs),
	}
}

func (o *ArchiveOptions) FromProto(x *proto.ArchiveRequest) {
	protoPaths := x.GetPaths()

	paths := make([]string, len(protoPaths))
	for i, p := range protoPaths {
		paths[i] = string(p)
	}

	*o = ArchiveOptions{
		Treeish: x.GetTreeish(),
		Format:  ArchiveFormatFromProto(x.GetFormat()),
		Paths:   paths,
	}
}

func (o *ArchiveOptions) ToProto(repo string) *proto.ArchiveRequest {
	paths := make([][]byte, len(o.Paths))
	for i, p := range o.Paths {
		paths[i] = []byte(p)
	}

	return &proto.ArchiveRequest{
		Repo:    repo,
		Treeish: o.Treeish,
		Format:  o.Format.ToProto(),
		Paths:   paths,
	}
}

type DiffOptions struct {
	// These fields must be valid <commit> inputs as defined by gitrevisions(7).
	Base string
	Head string

	// RangeType to be used for computing the diff: one of ".." or "..." (or unset: "").
	// For a nice visual explanation of ".." vs "...", see https://stackoverflow.com/a/46345364/2682729
	RangeType string

	// Paths specifies the paths to filter for during diffing.
	//
	// NOTE: Rename detection does not work if only the old path or the new path
	// is specified in this slice.
	Paths []string

	// InterHunkContext specifies the number of lines to consider for fusing hunks
	// together. I.e., when set to 5 and between 2 hunks there are at most 5 lines,
	// the 2 hunks will be fused together into a single chunk.
	//
	// The default for this is 3.
	InterHunkContext *int

	// ContextLines specifies the number of lines of context to show around added/removed
	// lines.
	// This is the number of lines that will be shown before and after each line that
	// has been added/removed. If InterHunkContext is not zero, the context will still
	// be fused together with other hunks if they meet the threshold.
	//
	// The default for this is 3.
	ContextLines *int
}

type Client interface {
	// Scoped adds a usage scope to the client and returns a new client with that scope.
	// Usage scopes should be descriptive and be lowercase plaintext, eg. batches.reconciler.
	// Scopes should get more specific as they get nested, eg:
	// batches.reconciler
	// batches.reconciler.processor
	// The Scoped method adds a single scope, so for gitserver.NewClient("batches").Scoped("reconciler")
	// the scope name will be batches.reconciler.
	// You may also use gitserver.NewClient("batches.reconciler") directly, where
	// useful.
	// We use scopes to add context to logs and metrics.
	Scoped(scope string) Client

	// AddrForRepo returns the gitserver address to use for the given repo name.
	AddrForRepo(ctx context.Context, repoName api.RepoName) string

	// ArchiveReader streams back the file contents of an archived git repo.
	ArchiveReader(ctx context.Context, repo api.RepoName, options ArchiveOptions) (io.ReadCloser, error)

	// StreamBlameFile returns Git blame information about a file in a streaming fashion.
	StreamBlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions) (HunkReader, error)

	// CreateCommitFromPatch will attempt to create a commit from a patch
	// If possible, the error returned will be of type protocol.CreateCommitFromPatchError
	CreateCommitFromPatch(context.Context, protocol.CreateCommitFromPatchRequest) (*protocol.CreateCommitFromPatchResponse, error)

	// GetDefaultBranch returns the name of the default branch and the commit it's
	// currently at from the given repository. If short is true, then `main` instead
	// of `refs/heads/main` would be returned.
	//
	// If the repository is empty or currently being cloned, empty values and no
	// error are returned.
	GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error)

	// GetObject fetches git object data in the supplied repo
	GetObject(ctx context.Context, repo api.RepoName, objectName string) (*gitdomain.GitObject, error)

	// IsRepoCloneable returns nil if the repository is cloneable.
	IsRepoCloneable(context.Context, api.RepoName) error

	// ListRefs returns a list of all refs in the repository.
	ListRefs(ctx context.Context, repo api.RepoName, opt ListRefsOpts) ([]gitdomain.Ref, error)

	// MergeBase returns the merge base commit sha for the specified revspecs.
	MergeBase(ctx context.Context, repo api.RepoName, base, head string) (api.CommitID, error)

	// MergeBaseOctopus returns the octopus merge base commit sha for the specified revspecs.
	MergeBaseOctopus(ctx context.Context, repo api.RepoName, revspecs ...string) (api.CommitID, error)

	RepoCloneProgress(context.Context, api.RepoName) (*protocol.RepoCloneProgress, error)

	// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
	// used.
	//
	// Error cases:
	// * Repo does not exist: gitdomain.RepoNotExistError
	// * Commit does not exist: gitdomain.RevisionNotFoundError
	// * Empty repository: gitdomain.RevisionNotFoundError
	// * Other unexpected errors.
	ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (api.CommitID, error)

	// RevAtTime returns the OID of the nearest ancestor of `spec` that has a
	// commit time before the given time. To simplify the logic, it only
	// follows the first parent of merge commits to linearize the commit
	// history. The intent is to return the state of a branch at a given time.
	//
	// If `spec` does not exist, an error will be returned. In the case
	// no commit in the ancestry of `spec` before the time `t`, no error
	// is returned, but the second return value `found` will be false.
	RevAtTime(ctx context.Context, repo api.RepoName, spec string, t time.Time) (oid api.CommitID, found bool, err error)

	// Search executes a search as specified by args, streaming the results as
	// it goes by calling onMatches with each set of results it receives in
	// response.
	Search(_ context.Context, _ *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (limitHit bool, _ error)

	// Stat returns a FileInfo describing the named file at commit.
	Stat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error)

	// ReadDir reads the contents of the named directory at commit.
	ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) (ReadDirIterator, error)

	// NewFileReader returns an io.ReadCloser reading from the named file at commit.
	// The caller should always close the reader after use.
	//
	// If you just need to check a file's existence, use Stat, not a file reader.
	//
	// If the file doesn't exist, the returned error will pass the os.IsNotExist()
	// check. Subrepo permissions are respected by this method and if no access
	// is granted, the error will also pass the os.IsNotExist() check.
	//
	// If the path points to a submodule, a reader for an empty file is returned
	// (ie. io.EOF is returned immediately).
	//
	// If the specified commit does not exist, a RevisionNotFoundError is returned.
	NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (io.ReadCloser, error)

	// ChangedFiles returns the list of files that have been added, modified, or
	// deleted in the entire repository between the two given <tree-ish> identifiers (e.g., commit, branch, tag).
	//
	// If base is omitted, the parent of head is used as the base.
	//
	// If either the base or head <tree-ish> id does not exist, a gitdomain.RevisionNotFoundError is returned.
	ChangedFiles(ctx context.Context, repo api.RepoName, base, head string) (ChangedFilesIterator, error)

	// Commits returns all commits matching the options.
	Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions) ([]*gitdomain.Commit, error)

	// FirstEverCommit returns the first commit ever made to the repository.
	FirstEverCommit(ctx context.Context, repo api.RepoName) (*gitdomain.Commit, error)

	// Diff returns an iterator that can be used to access the diff between two
	// commits on a per-file basis. The iterator must be closed with Close when no
	// longer required.
	Diff(ctx context.Context, repo api.RepoName, opts DiffOptions) (*DiffFileIterator, error)

	// GetCommit returns the commit with the given commit ID, or RevisionNotFoundError if no such commit
	// exists.
	GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID) (*gitdomain.Commit, error)

	// BehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
	// revspecs).
	BehindAhead(ctx context.Context, repo api.RepoName, left, right string) (*gitdomain.BehindAhead, error)

	// ContributorCount returns the number of commits grouped by contributor
	ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) ([]*gitdomain.ContributorCount, error)

	// SystemsInfo returns information about all gitserver instances associated with a Sourcegraph instance.
	SystemsInfo(ctx context.Context) ([]protocol.SystemInfo, error)

	// SystemInfo returns information about the gitserver instance at the given address.
	SystemInfo(ctx context.Context, addr string) (protocol.SystemInfo, error)

	// IsPerforcePathCloneable checks if the given Perforce depot path is cloneable by
	// checking if it is a valid depot and the given user has permission to access it.
	IsPerforcePathCloneable(ctx context.Context, conn protocol.PerforceConnectionDetails, depotPath string) error

	// CheckPerforceCredentials checks if the given Perforce credentials are valid
	CheckPerforceCredentials(ctx context.Context, conn protocol.PerforceConnectionDetails) error

	// PerforceUsers lists all the users known to the given Perforce server.
	PerforceUsers(ctx context.Context, conn protocol.PerforceConnectionDetails) ([]*perforce.User, error)

	// PerforceProtectsForUser returns all protects that apply to the given Perforce user.
	PerforceProtectsForUser(ctx context.Context, conn protocol.PerforceConnectionDetails, username string) ([]*perforce.Protect, error)

	// PerforceProtectsForDepot returns all protects that apply to the given Perforce depot.
	PerforceProtectsForDepot(ctx context.Context, conn protocol.PerforceConnectionDetails, depot string) ([]*perforce.Protect, error)

	// PerforceGroupMembers returns the members of the given Perforce group.
	PerforceGroupMembers(ctx context.Context, conn protocol.PerforceConnectionDetails, group string) ([]string, error)

	// IsPerforceSuperUser checks if the given Perforce user is a super user, and otherwise returns an error.
	IsPerforceSuperUser(ctx context.Context, conn protocol.PerforceConnectionDetails) error

	// PerforceGetChangelist gets the perforce changelist details for the given changelist ID.
	PerforceGetChangelist(ctx context.Context, conn protocol.PerforceConnectionDetails, changelist string) (*perforce.Changelist, error)

	// ListGitoliteRepos returns a list of Gitolite repositories. Gitserver owns the SSH keys
	// so is the only service able to talk to gitolite.
	ListGitoliteRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

func (c *clientImplementor) SystemsInfo(ctx context.Context) (_ []protocol.SystemInfo, err error) {
	ctx, _, endObservation := c.operations.systemsInfo.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
	})
	defer endObservation(1, observation.Args{})

	addresses := c.clientSource.Addresses()

	wg := pool.NewWithResults[protocol.SystemInfo]().WithErrors().WithContext(ctx)

	for _, addr := range addresses {
		addr := addr // capture addr
		wg.Go(func(ctx context.Context) (protocol.SystemInfo, error) {
			response, err := c.getDiskInfo(ctx, addr)
			if err != nil {
				return protocol.SystemInfo{}, err
			}
			return protocol.SystemInfo{
				Address:     addr.Address(),
				FreeSpace:   response.GetFreeSpace(),
				TotalSpace:  response.GetTotalSpace(),
				PercentUsed: response.GetPercentUsed(),
			}, nil
		})
	}

	return wg.Wait()
}

func (c *clientImplementor) SystemInfo(ctx context.Context, addr string) (_ protocol.SystemInfo, err error) {
	ctx, _, endObservation := c.operations.systemInfo.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("addr", addr),
		},
	})
	defer endObservation(1, observation.Args{})

	ac := c.clientSource.GetAddressWithClient(addr)
	if ac == nil {
		return protocol.SystemInfo{}, errors.Newf("no client for address: %s", addr)
	}

	response, err := c.getDiskInfo(ctx, ac)
	if err != nil {
		return protocol.SystemInfo{}, err
	}

	return protocol.SystemInfo{
		Address:    ac.Address(),
		FreeSpace:  response.FreeSpace,
		TotalSpace: response.TotalSpace,
	}, nil
}

func (c *clientImplementor) getDiskInfo(ctx context.Context, addr AddressWithClient) (*proto.DiskInfoResponse, error) {
	cli, err := addr.GRPCClient()
	if err != nil {
		return nil, err
	}

	resp, err := cli.DiskInfo(ctx, &proto.DiskInfoRequest{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *clientImplementor) AddrForRepo(ctx context.Context, repo api.RepoName) string {
	return c.clientSource.AddrForRepo(ctx, repo)
}

func clientForConn(conn *grpc.ClientConn) proto.GitserverServiceClient {
	return &errorTranslatingClient{
		base: &automaticRetryClient{
			base: proto.NewGitserverServiceClient(conn),
		},
	}
}

func (c *clientImplementor) Search(ctx context.Context, args *protocol.SearchRequest, onMatches func([]protocol.CommitMatch)) (_ bool, err error) {
	ctx, _, endObservation := c.operations.search.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			args.Repo.Attr(),
			attribute.Stringer("query", args.Query),
			attribute.Bool("diff", args.IncludeDiff),
			attribute.Int("limit", args.Limit),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, args.Repo)
	if err != nil {
		return false, err
	}

	cs, err := client.Search(ctx, args.ToProto())
	if err != nil {
		return false, err
	}

	limitHit := false
	for {
		msg, err := cs.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return limitHit, nil
			}
			return limitHit, err
		}

		switch m := msg.Message.(type) {
		case *proto.SearchResponse_LimitHit:
			limitHit = limitHit || m.LimitHit
		case *proto.SearchResponse_Match:
			onMatches([]protocol.CommitMatch{protocol.CommitMatchFromProto(m.Match)})
		default:
			return false, errors.Newf("unknown message type %T", m)
		}
	}
}

// MockIsRepoCloneable mocks (*Client).IsRepoCloneable for tests.
var MockIsRepoCloneable func(api.RepoName) error

func (c *clientImplementor) IsRepoCloneable(ctx context.Context, repo api.RepoName) (err error) {
	ctx, _, endObservation := c.operations.isRepoCloneable.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	if MockIsRepoCloneable != nil {
		return MockIsRepoCloneable(repo)
	}

	var resp protocol.IsRepoCloneableResponse

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return err
	}

	req := &proto.IsRepoCloneableRequest{
		Repo: string(repo),
	}

	r, err := client.IsRepoCloneable(ctx, req)
	if err != nil {
		return err
	}

	resp.FromProto(r)

	if resp.Cloneable {
		return nil
	}

	// Treat all 4xx errors as not found, since we have more relaxed
	// requirements on what a valid URL is we should treat bad requests,
	// etc as not found.
	notFound := strings.Contains(resp.Reason, "not found") || strings.Contains(resp.Reason, "The requested URL returned error: 4")
	return &RepoNotCloneableErr{
		repo:     repo,
		reason:   resp.Reason,
		notFound: notFound,
		cloned:   resp.Cloned,
	}
}

// RepoNotCloneableErr is the error that happens when a repository can not be cloned.
type RepoNotCloneableErr struct {
	repo     api.RepoName
	reason   string
	notFound bool
	cloned   bool // Has the repo ever been cloned in the past
}

// NotFound returns true if the repo could not be cloned because it wasn't found.
// This may be because the repo doesn't exist, or because the repo is private and
// there are insufficient permissions.
func (e *RepoNotCloneableErr) NotFound() bool {
	return e.notFound
}

func (e *RepoNotCloneableErr) Error() string {
	msg := "unable to clone repo"
	if e.cloned {
		msg = "unable to update repo"
	}
	return fmt.Sprintf("%s (name=%q notfound=%v) because %s", msg, e.repo, e.notFound, e.reason)
}

func (c *clientImplementor) RepoCloneProgress(ctx context.Context, repo api.RepoName) (_ *protocol.RepoCloneProgress, err error) {
	ctx, _, endObservation := c.operations.repoCloneProgress.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	res, err := client.RepoCloneProgress(ctx, &proto.RepoCloneProgressRequest{RepoName: string(repo)})
	if err != nil {
		return nil, err
	}

	var rcp protocol.RepoCloneProgress
	rcp.FromProto(res)

	return &rcp, nil
}

func (c *clientImplementor) IsPerforcePathCloneable(ctx context.Context, conn protocol.PerforceConnectionDetails, depotPath string) (err error) {
	ctx, _, endObservation := c.operations.isPerforcePathCloneable.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
	})
	defer endObservation(1, observation.Args{})

	// depotPath is not actually a repo name, but it will spread the load of isPerforcePathCloneable
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(depotPath))
	if err != nil {
		return err
	}
	_, err = client.IsPerforcePathCloneable(ctx, &proto.IsPerforcePathCloneableRequest{
		ConnectionDetails: conn.ToProto(),
		DepotPath:         depotPath,
	})
	if err != nil {
		// Unwrap proto errors for nicer error messages.
		if s, ok := status.FromError(err); ok {
			return errors.New(s.Message())
		}
		return err
	}

	return nil
}

func (c *clientImplementor) CheckPerforceCredentials(ctx context.Context, conn protocol.PerforceConnectionDetails) (err error) {
	ctx, _, endObservation := c.operations.checkPerforceCredentials.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return err
	}
	_, err = client.CheckPerforceCredentials(ctx, &proto.CheckPerforceCredentialsRequest{
		ConnectionDetails: conn.ToProto(),
	})
	if err != nil {
		// Unwrap proto errors for nicer error messages.
		if s, ok := status.FromError(err); ok {
			return errors.New(s.Message())
		}
		return err
	}

	return nil
}

func (c *clientImplementor) PerforceUsers(ctx context.Context, conn protocol.PerforceConnectionDetails) (_ []*perforce.User, err error) {
	ctx, _, endObservation := c.operations.perforceUsers.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return nil, err
	}
	resp, err := client.PerforceUsers(ctx, &proto.PerforceUsersRequest{
		ConnectionDetails: conn.ToProto(),
	})
	if err != nil {
		return nil, err
	}

	users := make([]*perforce.User, len(resp.GetUsers()))
	for i, u := range resp.GetUsers() {
		users[i] = perforce.UserFromProto(u)
	}
	return users, nil
}

func (c *clientImplementor) PerforceProtectsForUser(ctx context.Context, conn protocol.PerforceConnectionDetails, username string) (_ []*perforce.Protect, err error) {
	ctx, _, endObservation := c.operations.perforceProtectsForUser.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("username", username),
		},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return nil, err
	}
	resp, err := client.PerforceProtectsForUser(ctx, &proto.PerforceProtectsForUserRequest{
		ConnectionDetails: conn.ToProto(),
		Username:          username,
	})
	if err != nil {
		return nil, err
	}

	protects := make([]*perforce.Protect, len(resp.GetProtects()))
	for i, p := range resp.GetProtects() {
		protects[i] = perforce.ProtectFromProto(p)
	}
	return protects, nil
}

func (c *clientImplementor) PerforceProtectsForDepot(ctx context.Context, conn protocol.PerforceConnectionDetails, depot string) (_ []*perforce.Protect, err error) {
	ctx, _, endObservation := c.operations.perforceProtectsForDepot.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("depot", depot),
		},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return nil, err
	}
	resp, err := client.PerforceProtectsForDepot(ctx, &proto.PerforceProtectsForDepotRequest{
		ConnectionDetails: conn.ToProto(),
		Depot:             depot,
	})
	if err != nil {
		return nil, err
	}

	protects := make([]*perforce.Protect, len(resp.GetProtects()))
	for i, p := range resp.GetProtects() {
		protects[i] = perforce.ProtectFromProto(p)
	}
	return protects, nil
}

func (c *clientImplementor) PerforceGroupMembers(ctx context.Context, conn protocol.PerforceConnectionDetails, group string) (_ []string, err error) {
	ctx, _, endObservation := c.operations.perforceGroupMembers.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("group", group),
		},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return nil, err
	}
	resp, err := client.PerforceGroupMembers(ctx, &proto.PerforceGroupMembersRequest{
		ConnectionDetails: conn.ToProto(),
		Group:             group,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUsernames(), nil
}

func (c *clientImplementor) IsPerforceSuperUser(ctx context.Context, conn protocol.PerforceConnectionDetails) (err error) {
	ctx, _, endObservation := c.operations.isPerforceSuperUser.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return err
	}
	_, err = client.IsPerforceSuperUser(ctx, &proto.IsPerforceSuperUserRequest{
		ConnectionDetails: conn.ToProto(),
	})
	return err
}

func (c *clientImplementor) PerforceGetChangelist(ctx context.Context, conn protocol.PerforceConnectionDetails, changelist string) (_ *perforce.Changelist, err error) {
	ctx, _, endObservation := c.operations.perforceGetChangelist.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("changelist", changelist),
		},
	})
	defer endObservation(1, observation.Args{})

	// p4port is not actually a repo name, but it will spread the load of CheckPerforceCredentials
	// a bit over the different gitserver instances. It's really just used as a consistent hashing
	// key here.
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(conn.P4Port))
	if err != nil {
		return nil, err
	}
	resp, err := client.PerforceGetChangelist(ctx, &proto.PerforceGetChangelistRequest{
		ConnectionDetails: conn.ToProto(),
		ChangelistId:      changelist,
	})
	if err != nil {
		return nil, err
	}

	return perforce.ChangelistFromProto(resp.GetChangelist()), nil
}

func (c *clientImplementor) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (_ *protocol.CreateCommitFromPatchResponse, err error) {
	ctx, _, endObservation := c.operations.createCommitFromPatch.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			req.Repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, req.Repo)
	if err != nil {
		return nil, err
	}

	cc, err := client.CreateCommitFromPatchBinary(ctx)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			for _, detail := range st.Details() {
				switch dt := detail.(type) {
				case *proto.CreateCommitFromPatchError:
					var e protocol.CreateCommitFromPatchError
					e.FromProto(dt)
					return nil, &e
				}
			}
		}
		return nil, err
	}

	// Send the metadata event first.
	if err := cc.Send(&proto.CreateCommitFromPatchBinaryRequest{Payload: &proto.CreateCommitFromPatchBinaryRequest_Metadata_{
		Metadata: req.ToMetadataProto(),
	}}); err != nil {
		return nil, errors.Wrap(err, "sending metadata")
	}

	// Then create a writer that sends data in chunks that won't exceed the maximum
	// message size of gRPC of the patch in separate events.
	w := streamio.NewWriter(func(p []byte) error {
		req := &proto.CreateCommitFromPatchBinaryRequest{
			Payload: &proto.CreateCommitFromPatchBinaryRequest_Patch_{
				Patch: &proto.CreateCommitFromPatchBinaryRequest_Patch{
					Data: p,
				},
			},
		}
		return cc.Send(req)
	})

	if _, err := w.Write(req.Patch); err != nil {
		return nil, errors.Wrap(err, "writing chunk of patch")
	}

	resp, err := cc.CloseAndRecv()
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, err
		}

		for _, detail := range st.Details() {
			switch dt := detail.(type) {
			case *proto.CreateCommitFromPatchError:
				var e protocol.CreateCommitFromPatchError
				e.FromProto(dt)
				return nil, &e
			}
		}

		return nil, err
	}

	var res protocol.CreateCommitFromPatchResponse
	res.FromProto(resp, nil)

	return &res, nil
}

func (c *clientImplementor) GetObject(ctx context.Context, repo api.RepoName, objectName string) (_ *gitdomain.GitObject, err error) {
	ctx, _, endObservation := c.operations.getObject.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("name", objectName),
		},
	})
	defer endObservation(1, observation.Args{})

	req := protocol.GetObjectRequest{
		Repo:       repo,
		ObjectName: objectName,
	}
	client, err := c.clientSource.ClientForRepo(ctx, req.Repo)
	if err != nil {
		return nil, err
	}

	grpcResp, err := client.GetObject(ctx, req.ToProto())
	if err != nil {
		return nil, err
	}

	var res protocol.GetObjectResponse
	res.FromProto(grpcResp)

	return &res.Object, nil
}

func stringsToByteSlices(in []string) [][]byte {
	res := make([][]byte, len(in))
	for i, s := range in {
		res[i] = []byte(s)
	}
	return res
}

func byteSlicesToStrings(in [][]byte) []string {
	res := make([]string, len(in))
	for i, s := range in {
		res[i] = string(s)
	}
	return res
}

func (c *clientImplementor) ListGitoliteRepos(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	client, err := c.clientSource.ClientForRepo(ctx, api.RepoName(gitoliteHost))
	if err != nil {
		return nil, err
	}

	req := &proto.ListGitoliteRequest{
		GitoliteHost: gitoliteHost,
	}

	grpcResp, err := client.ListGitolite(ctx, req)
	if err != nil {
		return nil, err
	}

	list = make([]*gitolite.Repo, len(grpcResp.Repos))
	for i, r := range grpcResp.GetRepos() {
		list[i] = &gitolite.Repo{
			Name: r.GetName(),
			URL:  r.GetUrl(),
		}
	}

	return list, nil
}

// ReadDirIterator is an iterator for retrieving the file descriptors of files/dirs
// in a Git repository.
//
// The caller must ensure that they call Close() when the iterator is no longer
// needed to release any associated resources.
type ReadDirIterator interface {
	// Next returns the next file descriptor.
	//
	// If there are no more files, Next returns an io.EOF error.
	// If an error occurs during iteration, Next returns the error that occurred.
	Next() (fs.FileInfo, error)

	// Close closes the iterator and releases any associated resources.
	//
	// It is important to call Close when the iterator is no longer needed to avoid resource leaks.
	// After calling Close, any subsequent calls to Next will return an io.EOF error.
	Close()
}

// ChangedFilesIterator is an iterator for retrieving the status of changed files in a Git repository.
// It allows iterating over the changed files and retrieving their paths and statuses.
//
// The caller must ensure that they call Close() when the iterator is no longer needed to release any associated resources.
type ChangedFilesIterator interface {
	// Next returns the next changed file's path and status.
	//
	// If there are no more changed files, Next returns an io.EOF error.
	// If an error occurs during iteration, Next returns the error that occurred.
	Next() (gitdomain.PathStatus, error)

	// Close closes the iterator and releases any associated resources.
	//
	// It is important to call Close when the iterator is no longer needed to avoid resource leaks.
	// After calling Close, any subsequent calls to Next will return an io.EOF error.
	Close()
}

// NewChangedFilesIteratorFromSlice returns a new ChangedFilesIterator that iterates over the given slice of changed files (in order),
// which is useful for testing.
func NewChangedFilesIteratorFromSlice(files []gitdomain.PathStatus) ChangedFilesIterator {
	return &changedFilesSliceIterator{files: files}
}

type changedFilesSliceIterator struct {
	mu     sync.Mutex
	files  []gitdomain.PathStatus
	closed bool
}

func (c *changedFilesSliceIterator) Next() (gitdomain.PathStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return gitdomain.PathStatus{}, io.EOF
	}

	if len(c.files) == 0 {
		return gitdomain.PathStatus{}, io.EOF
	}

	file := c.files[0]
	c.files = c.files[1:]

	return file, nil
}

func (c *changedFilesSliceIterator) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
}

var _ ChangedFilesIterator = &changedFilesSliceIterator{}

// NewReadDirIteratorFromSlice returns a new ReadDirIterator that iterates over
// the given slice which is useful for testing.
func NewReadDirIteratorFromSlice(fds []fs.FileInfo) ReadDirIterator {
	return &readDirSliceIterator{fds: fds}
}

type readDirSliceIterator struct {
	mu     sync.Mutex
	fds    []fs.FileInfo
	closed bool
}

func (c *readDirSliceIterator) Next() (fs.FileInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, io.EOF
	}

	if len(c.fds) == 0 {
		return nil, io.EOF
	}

	fd := c.fds[0]
	c.fds = c.fds[1:]

	return fd, nil
}

func (c *readDirSliceIterator) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
}

var _ ReadDirIterator = &readDirSliceIterator{}
