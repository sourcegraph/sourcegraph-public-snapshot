package gitserver

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Diff returns an iterator that can be used to access the diff between two
// commits on a per-file basis. The iterator must be closed with Close when no
// longer required.
func (c *clientImplementor) Diff(ctx context.Context, repo api.RepoName, opts DiffOptions) (_ *DiffFileIterator, err error) {
	ctx, _, endObservation := c.operations.diff.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})

	req := &proto.RawDiffRequest{
		RepoName:    string(repo),
		BaseRevSpec: []byte(opts.Base),
		HeadRevSpec: []byte(opts.Head),
		Paths:       stringsToByteSlices(opts.Paths),
	}

	if opts.InterHunkContext != nil {
		req.InterHunkContext = pointers.Ptr(uint32(*opts.InterHunkContext))
	}

	if opts.ContextLines != nil {
		req.ContextLines = pointers.Ptr(uint32(*opts.ContextLines))
	}

	// Rare case: the base is the empty tree, in which case we must use ..
	// instead of ... as the latter only works for commits.
	if opts.Base == DevNullSHA {
		req.ComparisonType = proto.RawDiffRequest_COMPARISON_TYPE_ONLY_IN_HEAD
	} else if opts.RangeType != ".." {
		req.ComparisonType = proto.RawDiffRequest_COMPARISON_TYPE_INTERSECTION
	} else {
		req.ComparisonType = proto.RawDiffRequest_COMPARISON_TYPE_ONLY_IN_HEAD
	}

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	cc, err := client.RawDiff(ctx, req)
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We start by reading the first message to early-exit on potential errors.
	firstResp, firstRespErr := cc.Recv()
	if firstRespErr != nil {
		if errors.HasType[*gitdomain.RevisionNotFoundError](firstRespErr) {
			cancel()
			err = firstRespErr
			endObservation(1, observation.Args{})
			return nil, err
		}
	}

	firstRespRead := false
	r := streamio.NewReader(func() ([]byte, error) {
		if !firstRespRead {
			firstRespRead = true
			if firstRespErr != nil {
				return nil, firstRespErr
			}
			return firstResp.GetChunk(), nil
		}

		m, err := cc.Recv()
		if err != nil {
			return nil, err
		}
		return m.GetChunk(), nil
	})

	return &DiffFileIterator{
		onClose: func() {
			cancel()
			endObservation(1, observation.Args{})

		},
		mfdr:           diff.NewMultiFileDiffReader(r),
		fileFilterFunc: getFilterFunc(ctx, c.subRepoPermsChecker, repo),
	}, nil
}

type DiffFileIterator struct {
	onClose        func()
	mfdr           *diff.MultiFileDiffReader
	fileFilterFunc diffFileIteratorFilter
}

func NewDiffFileIterator(rdr io.ReadCloser) *DiffFileIterator {
	return &DiffFileIterator{
		mfdr: diff.NewMultiFileDiffReader(rdr),
	}
}

type diffFileIteratorFilter func(fileName string) (bool, error)

func getFilterFunc(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName) diffFileIteratorFilter {
	if !authz.SubRepoEnabled(checker) {
		return nil
	}
	return func(fileName string) (bool, error) {
		shouldFilter, err := authz.FilterActorPath(ctx, checker, actor.FromContext(ctx), repo, fileName)
		if err != nil {
			return false, err
		}
		return shouldFilter, nil
	}
}

func (i *DiffFileIterator) Close() error {
	if i.onClose != nil {
		i.onClose()
	}
	return nil
}

// Next returns the next file diff. If no more diffs are available, the diff
// will be nil and the error will be io.EOF.
func (i *DiffFileIterator) Next() (*diff.FileDiff, error) {
	fd, err := i.mfdr.ReadFile()
	if err != nil {
		return fd, err
	}
	if i.fileFilterFunc != nil {
		if canRead, err := i.fileFilterFunc(fd.NewName); err != nil {
			return nil, err
		} else if !canRead {
			// go to next
			return i.Next()
		}
	}
	return fd, err
}

// ContributorOptions contains options for filtering contributor commit counts
type ContributorOptions struct {
	Range string    // the range for which stats will be fetched
	After time.Time // the date after which to collect commits
	Path  string    // compute stats for commits that touch this path
}

func (o *ContributorOptions) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("range", o.Range),
		attribute.String("after", o.After.Format(time.RFC3339)),
		attribute.String("path", o.Path),
	}
}

func (c *clientImplementor) ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) (_ []*gitdomain.ContributorCount, err error) {
	ctx, _, endObservation := c.operations.contributorCount.With(ctx,
		&err,
		observation.Args{
			Attrs:             opt.Attrs(),
			MetricLabelValues: []string{c.scope},
		},
	)
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	res, err := client.ContributorCounts(ctx, &proto.ContributorCountsRequest{
		RepoName: string(repo),
		Range:    []byte(opt.Range),
		After:    timestamppb.New(opt.After),
		Path:     []byte(opt.Path),
	})
	if err != nil {
		return nil, err
	}

	counts := make([]*gitdomain.ContributorCount, len(res.GetCounts()))
	for i, c := range res.GetCounts() {
		counts[i] = gitdomain.ContributorCountFromProto(c)
	}

	return counts, nil
}

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func (c *clientImplementor) ChangedFiles(ctx context.Context, repo api.RepoName, base, head string) (iterator ChangedFilesIterator, err error) {
	ctx, _, endObservation := c.operations.changedFiles.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("base", base),
			attribute.String("head", head),
		},
	})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := client.ChangedFiles(ctx, &proto.ChangedFilesRequest{
		RepoName: string(repo),
		Base:     []byte(base),
		Head:     []byte(head),
	})
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	fetchFunc := func() ([]gitdomain.PathStatus, error) {
		resp, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		protoFiles := resp.GetFiles()

		changes := make([]gitdomain.PathStatus, 0, len(protoFiles))
		for _, p := range protoFiles {
			changes = append(changes, gitdomain.PathStatusFromProto(p))
		}

		return changes, nil
	}

	closeFunc := func() {
		cancel()
		endObservation(1, observation.Args{})
	}

	return newChangedFilesIterator(fetchFunc, closeFunc), nil
}

func newChangedFilesIterator(fetchFunc func() ([]gitdomain.PathStatus, error), closeFunc func()) *changedFilesIterator {
	return &changedFilesIterator{
		fetchFunc: fetchFunc,
		closeFunc: closeFunc,
		closeChan: make(chan struct{}),
	}
}

type changedFilesIterator struct {
	// fetchFunc is the function that will be invoked when the buffer is empty.
	//
	// The function should return the next batch of data, or an error if the fetch
	// failed.
	//
	// fetchFunc should return an io.EOF error when there is no more data to fetch.
	fetchFunc func() ([]gitdomain.PathStatus, error)
	fetchErr  error

	closeOnce sync.Once
	closeFunc func()
	closeChan chan struct{}

	buffer []gitdomain.PathStatus
}

func (i *changedFilesIterator) Next() (gitdomain.PathStatus, error) {
	select {
	case <-i.closeChan:
		return gitdomain.PathStatus{}, io.EOF
	default:
	}

	if i.fetchErr != nil {
		return gitdomain.PathStatus{}, i.fetchErr
	}

	for len(i.buffer) == 0 { // If we've exhausted the buffer, fetch more data
		// If we've exhausted the buffer, fetch more data
		//
		// We keep trying until we get a non-empty buffer since it's technically possible for the fetchFunc to return an empty slice
		// even if there is more data to fetch.

		i.buffer, i.fetchErr = i.fetchFunc()
		if i.fetchErr != nil { // Check if there was an error fetching the data
			return gitdomain.PathStatus{}, i.fetchErr
		}
	}

	out := i.buffer[0]
	i.buffer = i.buffer[1:]

	return out, nil
}

func (i *changedFilesIterator) Close() {
	i.closeOnce.Do(func() {
		if i.closeFunc != nil {
			i.closeFunc()
		}
		close(i.closeChan)
	})
}

func (c *clientImplementor) ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) (_ ReadDirIterator, err error) {
	ctx, _, endObservation := c.operations.readDir.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			commit.Attr(),
			attribute.String("path", path),
			attribute.Bool("recurse", recurse),
		},
	})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We create a context here to have a way to cancel the RPC if the caller
	// doesn't read the entire result.
	ctx, cancel := context.WithCancel(ctx)
	cc, err := client.ReadDir(ctx, &proto.ReadDirRequest{
		RepoName:  string(repo),
		CommitSha: string(commit),
		Path:      []byte(path),
		Recursive: recurse,
	})
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We receive the first chunk here so that we can return an error if the
	// file/path is not found.
	firstChunk, err := cc.Recv()
	if err != nil {
		defer endObservation(1, observation.Args{})
		defer cancel()

		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.NotFound {
				for _, d := range s.Details() {
					fp, ok := d.(*proto.FileNotFoundPayload)
					if ok {
						return nil, &os.PathError{Op: "open", Path: string(fp.Path), Err: os.ErrNotExist}
					}
				}
			}
		}

		return nil, err
	}

	return &readDirIterator{
		ctx:        ctx,
		firstChunk: firstChunk,
		cc:         cc,
		onClose: func() {
			cancel()
			endObservation(1, observation.Args{})
		},
		subRepoPermsChecker: c.subRepoPermsChecker,
		repo:                repo,
	}, nil
}

type readDirIterator struct {
	ctx                 context.Context
	firstChunk          *proto.ReadDirResponse
	firstChunkConsumed  bool
	cc                  proto.GitserverService_ReadDirClient
	onClose             func()
	subRepoPermsChecker authz.SubRepoPermissionChecker
	repo                api.RepoName
	buffer              []fs.FileInfo
}

func (i *readDirIterator) Next() (fs.FileInfo, error) {
	if len(i.buffer) == 0 {
		for {
			if i.ctx.Err() != nil {
				return nil, i.ctx.Err()
			}

			chunk, err := i.fetchChunk()
			if err != nil {
				return nil, err
			}
			fds := []fs.FileInfo{}
			for _, f := range chunk.GetFileInfo() {
				fds = append(fds, gitdomain.ProtoFileInfoToFS(f))
			}
			if authz.SubRepoEnabled(i.subRepoPermsChecker) {
				a := actor.FromContext(i.ctx)
				filtered, filteringErr := authz.FilterActorFileInfos(i.ctx, i.subRepoPermsChecker, a, i.repo, fds)
				if filteringErr != nil {
					return nil, errors.Wrap(err, "filtering paths")
				}
				i.buffer = filtered
			} else {
				i.buffer = fds
			}
			if len(i.buffer) > 0 {
				break
			}
		}
	}

	if len(i.buffer) == 0 {
		return nil, io.EOF
	}

	fd := i.buffer[0]
	i.buffer = i.buffer[1:]

	return fd, nil
}

func (i *readDirIterator) fetchChunk() (*proto.ReadDirResponse, error) {
	if !i.firstChunkConsumed {
		i.firstChunkConsumed = true
		return i.firstChunk, nil
	}
	return i.cc.Recv()
}

func (i *readDirIterator) Close() {
	i.onClose()
}

// StreamBlameFile returns Git blame information about a file.
func (c *clientImplementor) StreamBlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions) (_ HunkReader, err error) {
	ctx, _, endObservation := c.operations.streamBlameFile.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append([]attribute.KeyValue{
			repo.Attr(),
			attribute.String("path", path),
		}, opt.Attrs()...),
	})

	// First, verify that the actor has access to the given path.
	hasAccess, err := authz.FilterActorPath(ctx, c.subRepoPermsChecker, actor.FromContext(ctx), repo, path)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, fs.ErrNotExist
	}

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	req := &proto.BlameRequest{
		RepoName:         string(repo),
		Commit:           string(opt.NewestCommit),
		Path:             []byte(path), // The file path might not be utf-8 encoded.
		IgnoreWhitespace: opt.IgnoreWhitespace,
	}
	if opt.Range != nil {
		req.Range = &proto.BlameRange{
			StartLine: uint32(opt.Range.StartLine),
			EndLine:   uint32(opt.Range.EndLine),
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	cli, err := client.Blame(ctx, req)
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We start by reading the first hunk to early-exit on potential errors,
	// ie. permission denied errors or invalid git command.
	firstHunkResp, err := cli.Recv()
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			if s.Code() == codes.NotFound {
				for _, d := range s.Details() {
					switch d.(type) {
					case *proto.FileNotFoundPayload:
						cancel()
						endObservation(1, observation.Args{})
						return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
					}
				}
			}
		}

		if err != io.EOF {
			cancel()
			endObservation(1, observation.Args{})
			return nil, err
		}
	}

	var hunk *proto.BlameHunk
	if firstHunkResp != nil {
		hunk = firstHunkResp.GetHunk()
	}
	return &grpcBlameHunkReader{
		firstHunk:      hunk,
		firstHunkErr:   err,
		c:              cli,
		cancel:         cancel,
		endObservation: func() { endObservation(1, observation.Args{}) },
	}, nil
}

type grpcBlameHunkReader struct {
	firstHunk      *proto.BlameHunk
	firstHunkErr   error
	firstHunkRead  bool
	c              proto.GitserverService_BlameClient
	cancel         context.CancelFunc
	endObservation func()
}

func (r *grpcBlameHunkReader) Read() (_ *gitdomain.Hunk, err error) {
	if !r.firstHunkRead {
		r.firstHunkRead = true
		if r.firstHunkErr != nil {
			return nil, r.firstHunkErr
		}
		return gitdomain.HunkFromBlameProto(r.firstHunk), nil
	}
	p, err := r.c.Recv()
	if err != nil {
		return nil, err
	}
	return gitdomain.HunkFromBlameProto(p.GetHunk()), nil
}

func (r *grpcBlameHunkReader) Close() error {
	r.cancel()
	r.endObservation()
	return nil
}

// ResolveRevisionOptions configure how we resolve revisions.
// The zero value should contain appropriate default values.
type ResolveRevisionOptions struct {
	// If set, try to fetch from remote if revision doesn't exist locally.
	EnsureRevision bool
}

// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
// used.
//
// Error cases:
// * Repo does not exist: gitdomain.RepoNotExistError
// * Commit does not exist: gitdomain.RevisionNotFoundError
// * Empty repository: gitdomain.RevisionNotFoundError
// * Other unexpected errors.
func (c *clientImplementor) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (_ api.CommitID, err error) {
	ctx, _, endObservation := c.operations.resolveRevision.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("spec", spec),
			attribute.Bool("ensureRevision", opt.EnsureRevision),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return "", err
	}

	req := &proto.ResolveRevisionRequest{
		RepoName: string(repo),
		RevSpec:  []byte(spec),
	}
	if opt.EnsureRevision {
		req.EnsureRevision = pointers.Ptr(true)
	}
	res, err := client.ResolveRevision(ctx, req)
	if err != nil {
		return "", err
	}

	return api.CommitID(res.GetCommitSha()), nil
}

func (c *clientImplementor) RevAtTime(ctx context.Context, repo api.RepoName, spec string, date time.Time) (_ api.CommitID, ok bool, err error) {
	ctx, _, endObservation := c.operations.revAtTime.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("spec", spec),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return "", false, err
	}

	req := &proto.RevAtTimeRequest{
		RepoName: string(repo),
		RevSpec:  []byte(spec),
		Time:     timestamppb.New(date),
	}
	res, err := client.RevAtTime(ctx, req)
	if err != nil {
		return "", false, err
	}

	return api.CommitID(res.GetCommitSha()), res.GetCommitSha() != "", nil
}

func (c *clientImplementor) GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	ctx, _, endObservation := c.operations.getDefaultBranch.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return "", "", err
	}

	res, err := client.DefaultBranch(ctx, &proto.DefaultBranchRequest{
		RepoName: string(repo),
	})
	if err != nil {
		// If we fail to get the default branch due to cloning or being empty, we return nothing.
		if errors.HasType[*gitdomain.RepoNotExistError](err) {
			return "", "", nil
		}
		if errors.HasType[*gitdomain.RevisionNotFoundError](err) {
			return "", "", nil
		}
		return "", "", err
	}

	return res.GetRefName(), api.CommitID(res.GetCommit()), nil
}

func (c *clientImplementor) MergeBase(ctx context.Context, repo api.RepoName, base, head string) (_ api.CommitID, err error) {
	ctx, _, endObservation := c.operations.mergeBase.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("base", base),
			attribute.String("head", head),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return "", err
	}

	res, err := client.MergeBase(ctx, &proto.MergeBaseRequest{
		RepoName: string(repo),
		Base:     []byte(base),
		Head:     []byte(head),
	})
	if err != nil {
		return "", err
	}

	return api.CommitID(res.GetMergeBaseCommitSha()), nil
}

func (c *clientImplementor) MergeBaseOctopus(ctx context.Context, repo api.RepoName, revspecs ...string) (_ api.CommitID, err error) {
	ctx, _, endObservation := c.operations.mergeBaseOctopus.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.StringSlice("revspecs", revspecs),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return "", err
	}

	res, err := client.MergeBaseOctopus(ctx, &proto.MergeBaseOctopusRequest{
		RepoName: string(repo),
		Revspecs: stringsToByteSlices(revspecs),
	})
	if err != nil {
		return "", err
	}

	return api.CommitID(res.GetMergeBaseCommitSha()), nil
}

// BehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
// revspecs).
func (c *clientImplementor) BehindAhead(ctx context.Context, repo api.RepoName, left, right string) (_ *gitdomain.BehindAhead, err error) {
	ctx, _, endObservation := c.operations.behindAhead.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("left", left),
			attribute.String("right", right),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	response, err := client.BehindAhead(ctx, &proto.BehindAheadRequest{
		RepoName: string(repo),
		Left:     []byte(left),
		Right:    []byte(right),
	})
	if err != nil {
		return nil, err
	}

	return gitdomain.BehindAheadFromProto(response), nil
}

func (c *clientImplementor) NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (_ io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.newFileReader.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			commit.Attr(),
			attribute.String("name", name),
		},
	})

	// First, verify the actor can see the path.
	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, c.subRepoPermsChecker, a, repo, name); err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, os.ErrNotExist
	}

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	req := &proto.ReadFileRequest{
		RepoName: string(repo),
		Commit:   string(commit),
		Path:     []byte(rel(name)),
	}

	ctx, cancel := context.WithCancel(ctx)
	cli, err := client.ReadFile(ctx, req)
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We start by reading the first message to early-exit on potential errors,
	// ie. permission denied errors or invalid git command.
	firstResp, firstRespErr := cli.Recv()
	if firstRespErr != nil {
		if s, ok := status.FromError(firstRespErr); ok {
			if s.Code() == codes.NotFound {
				for _, d := range s.Details() {
					switch d.(type) {
					case *proto.FileNotFoundPayload:
						cancel()
						err = firstRespErr
						endObservation(1, observation.Args{})
						return nil, &os.PathError{Op: "open", Path: string(req.GetPath()), Err: os.ErrNotExist}
					}
				}
			}
		}
		if errors.HasType[*gitdomain.RevisionNotFoundError](firstRespErr) {
			cancel()
			err = firstRespErr
			endObservation(1, observation.Args{})
			return nil, err
		}
	}

	firstRespRead := false
	r := streamio.NewReader(func() ([]byte, error) {
		if !firstRespRead {
			firstRespRead = true
			if firstRespErr != nil {
				return nil, firstRespErr
			}
			return firstResp.GetData(), nil
		}

		m, err := cli.Recv()
		if err != nil {
			return nil, err
		}
		return m.GetData(), nil
	})

	return &blobReader{
		Reader: r,
		onClose: func() {
			cancel()
			endObservation(1, observation.Args{})
		},
	}, nil
}

type blobReader struct {
	io.Reader
	onClose func()
}

func (br *blobReader) Close() error {
	br.onClose()
	return nil
}

// Stat returns a FileInfo describing the named file at commit.
func (c *clientImplementor) Stat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (_ fs.FileInfo, err error) {
	ctx, _, endObservation := c.operations.stat.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			commit.Attr(),
			attribute.String("path", path),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	res, err := client.Stat(ctx, &proto.StatRequest{
		RepoName:  string(repo),
		CommitSha: string(commit),
		Path:      []byte(path),
	})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			// If sub repo permissions deny access to the file, we return os.ErrNotExist.
			if s.Code() == codes.NotFound {
				for _, d := range s.Details() {
					fp, ok := d.(*proto.FileNotFoundPayload)
					if ok {
						return nil, &os.PathError{Op: "open", Path: string(fp.Path), Err: os.ErrNotExist}
					}
				}
			}
		}

		return nil, err
	}

	fi := gitdomain.ProtoFileInfoToFS(res.GetFileInfo())

	if !authz.SubRepoEnabled(c.subRepoPermsChecker) {
		return fi, nil
	}

	// Applying sub-repo permissions
	a := actor.FromContext(ctx)
	include, filteringErr := authz.FilterActorFileInfo(ctx, c.subRepoPermsChecker, a, repo, fi)
	if include && filteringErr == nil {
		return fi, nil
	} else {
		if filteringErr != nil {
			err = errors.Wrap(filteringErr, "filtering paths")
		} else {
			err = &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
		}
		return nil, err
	}
}

type CommitsOrder int

const (
	// CommitsOrderDefault uses the default ordering of git log: in reverse chronological order.
	// See https://git-scm.com/docs/git-log#_commit_ordering for more details.
	CommitsOrderDefault CommitsOrder = iota
	// Show no parents before all of its children are shown, but otherwise show commits
	// in the commit timestamp order.
	// See https://git-scm.com/docs/git-log#_commit_ordering for more details.
	CommitsOrderCommitDate
	// Show no parents before all of its children are shown, and avoid showing commits
	// on multiple lines of history intermixed.
	// See https://git-scm.com/docs/git-log#_commit_ordering for more details.
	CommitsOrderTopoDate
)

// CommitsOptions specifies options for Commits.
type CommitsOptions struct {
	AllRefs bool // if true, all refs are searched for commits, not just a given range. When set, Range should not be specified.
	// commit ranges to inspect (revspec, "A..B", "A...B", etc.).
	Ranges []string

	N    uint // limit the number of returned commits to this many (0 means no limit)
	Skip uint // skip this many commits at the beginning

	MessageQuery string // include only commits whose commit message contains this substring

	Author string    // include only commits whose author matches this
	After  time.Time // include only commits after this date
	Before time.Time // include only commits before this date

	Order CommitsOrder

	Path string // only commits modifying the given path are selected (optional)

	Follow bool // follow the history of the path beyond renames (works only for a single path)

	// When finding commits to include, follow only the first parent commit upon
	// seeing a merge commit. This option can give a better overview when viewing
	// the evolution of a particular topic branch, because merges into a topic
	// branch tend to be only about adjusting to updated upstream from time to time,
	// and this option allows you to ignore the individual commits brought in to
	// your history by such a merge.
	FirstParent bool

	// When true return the names of the files changed in the commit
	IncludeFiles bool
}

func (opt CommitsOptions) ToProto(repoName api.RepoName) (*proto.CommitLogRequest, error) {
	if len(opt.Ranges) > 0 && opt.AllRefs {
		return nil, errors.New("cannot specify both a range and AllRefs")
	}
	if len(opt.Ranges) == 0 && !opt.AllRefs {
		return nil, errors.New("must specify a range or AllRefs")
	}

	p := &proto.CommitLogRequest{
		RepoName:              string(repoName),
		MaxCommits:            uint32(opt.N),
		Skip:                  uint32(opt.Skip),
		FollowOnlyFirstParent: opt.FirstParent,
		MessageQuery:          []byte(opt.MessageQuery),
		AuthorQuery:           []byte(opt.Author),
		Path:                  []byte(opt.Path),
		FollowPathRenames:     opt.Follow,
		IncludeModifiedFiles:  opt.IncludeFiles,
		AllRefs:               opt.AllRefs,
		After:                 timestamppb.New(opt.After),
		Before:                timestamppb.New(opt.Before),
	}

	for _, r := range opt.Ranges {
		p.Ranges = append(p.Ranges, []byte(r))
	}

	switch opt.Order {
	case CommitsOrderCommitDate:
		p.Order = proto.CommitLogRequest_COMMIT_LOG_ORDER_COMMIT_DATE
	case CommitsOrderTopoDate:
		p.Order = proto.CommitLogRequest_COMMIT_LOG_ORDER_TOPO_DATE
	case CommitsOrderDefault:
		p.Order = proto.CommitLogRequest_COMMIT_LOG_ORDER_UNSPECIFIED
	default:
		return nil, errors.Newf("invalid ordering %d", opt.Order)
	}

	return p, nil
}

func (c *clientImplementor) GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID) (_ *gitdomain.Commit, err error) {
	ctx, _, endObservation := c.operations.getCommit.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			id.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	subRepoEnabled := false
	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		subRepoEnabled, err = authz.SubRepoEnabledForRepo(ctx, c.subRepoPermsChecker, repo)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check sub repo permissions")
		}
	}

	res, err := client.GetCommit(ctx, &proto.GetCommitRequest{
		RepoName:             string(repo),
		Commit:               string(id),
		IncludeModifiedFiles: subRepoEnabled,
	})
	if err != nil {
		return nil, err
	}

	commit := gitdomain.CommitFromProto(res.GetCommit())

	if subRepoEnabled {
		ok, err := hasAccessToCommit(ctx, &wrappedCommit{Commit: commit, files: byteSlicesToStrings(res.GetModifiedFiles())}, repo, c.subRepoPermsChecker)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check sub repo permissions")
		}
		if !ok {
			return nil, &gitdomain.RevisionNotFoundError{Repo: repo, Spec: string(id)}
		}
	}

	return commit, nil
}

// Commits returns all commits matching the options.
func (c *clientImplementor) Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions) (_ []*gitdomain.Commit, err error) {
	opt = maybeIncludeFiles(opt, c.subRepoPermsChecker)
	ctx, _, endObservation := c.operations.commits.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("opts", fmt.Sprintf("%#v", opt)),
		},
	})
	defer endObservation(1, observation.Args{})

	wrappedCommits, err := c.getWrappedCommits(ctx, repo, opt)
	if err != nil {
		return nil, err
	}

	filtered, err := filterCommits(ctx, c.subRepoPermsChecker, wrappedCommits, repo)
	if err != nil {
		return nil, errors.Wrap(err, "filtering commits")
	}

	if needMoreCommits(filtered, wrappedCommits, opt, c.subRepoPermsChecker) {
		return c.getMoreCommits(ctx, repo, opt, filtered)
	}
	return filtered, err
}

func filterCommits(ctx context.Context, checker authz.SubRepoPermissionChecker, commits []*wrappedCommit, repoName api.RepoName) ([]*gitdomain.Commit, error) {
	if !authz.SubRepoEnabled(checker) {
		return unWrapCommits(commits), nil
	}
	filtered := make([]*gitdomain.Commit, 0, len(commits))
	for _, commit := range commits {
		if hasAccess, err := hasAccessToCommit(ctx, commit, repoName, checker); hasAccess {
			filtered = append(filtered, commit.Commit)
		} else if err != nil {
			return nil, err
		}
	}
	return filtered, nil
}

func unWrapCommits(wrappedCommits []*wrappedCommit) []*gitdomain.Commit {
	commits := make([]*gitdomain.Commit, 0, len(wrappedCommits))
	for _, wc := range wrappedCommits {
		commits = append(commits, wc.Commit)
	}
	return commits
}

func hasAccessToCommit(ctx context.Context, commit *wrappedCommit, repoName api.RepoName, checker authz.SubRepoPermissionChecker) (bool, error) {
	a := actor.FromContext(ctx)
	if commit.files == nil || len(commit.files) == 0 {
		return true, nil // If commit has no files, assume user has access to view the commit.
	}
	for _, fileName := range commit.files {
		if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repoName, fileName); err != nil {
			return false, err
		} else if hasAccess {
			// if the user has access to one file modified in the commit, they have access to view the commit
			return true, nil
		}
	}
	return false, nil
}

func (c *clientImplementor) getWrappedCommits(ctx context.Context, repo api.RepoName, opt CommitsOptions) ([]*wrappedCommit, error) {
	cli, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	popt, err := opt.ToProto(repo)
	if err != nil {
		return nil, err
	}

	cc, err := cli.CommitLog(ctx, popt)
	if err != nil {
		return nil, err
	}

	wrappedCommits := []*wrappedCommit{}
	for {
		chunk, err := cc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		for _, c := range chunk.GetCommits() {
			commit := c.GetCommit()
			wrappedCommits = append(wrappedCommits, &wrappedCommit{
				Commit: gitdomain.CommitFromProto(commit),
				files:  byteSlicesToStrings(c.GetModifiedFiles()),
			})
		}
	}

	return wrappedCommits, nil
}

func needMoreCommits(filtered []*gitdomain.Commit, commits []*wrappedCommit, opt CommitsOptions, checker authz.SubRepoPermissionChecker) bool {
	if !authz.SubRepoEnabled(checker) {
		return false
	}
	if opt.N == 0 || isRequestForSingleCommit(opt) {
		return false
	}
	if len(filtered) < len(commits) {
		return true
	}
	return false
}

func isRequestForSingleCommit(opt CommitsOptions) bool {
	return len(opt.Ranges) == 1 && opt.N == 1
}

// getMoreCommits handles the case where a specific number of commits was requested via CommitsOptions, but after sub-repo
// filtering, fewer than that requested number was left. This function requests the next N commits (where N was the number
// originally requested), filters the commits, and determines if this is at least N commits total after filtering. If not,
// the loop continues until N total filtered commits are collected _or_ there are no commits left to request.
func (c *clientImplementor) getMoreCommits(ctx context.Context, repo api.RepoName, opt CommitsOptions, baselineCommits []*gitdomain.Commit) ([]*gitdomain.Commit, error) {
	// We want to place an upper bound on the number of times we loop here so that we
	// don't hit pathological conditions where a lot of filtering has been applied.
	const maxIterations = 5

	totalCommits := make([]*gitdomain.Commit, 0, opt.N)
	for range maxIterations {
		if uint(len(totalCommits)) == opt.N {
			break
		}
		// Increment the Skip number to get the next N commits
		opt.Skip += opt.N
		wrappedCommits, err := c.getWrappedCommits(ctx, repo, opt)
		if err != nil {
			return nil, err
		}
		filtered, err := filterCommits(ctx, c.subRepoPermsChecker, wrappedCommits, repo)
		if err != nil {
			return nil, err
		}
		// join the new (filtered) commits with those already fetched (potentially truncating the list to have length N if necessary)
		totalCommits = joinCommits(baselineCommits, filtered, opt.N)
		baselineCommits = totalCommits
		if uint(len(wrappedCommits)) < opt.N {
			// No more commits available before filtering, so return current total commits (e.g. the last "page" of N commits has been reached)
			break
		}
	}
	return totalCommits, nil
}

func joinCommits(previous, next []*gitdomain.Commit, desiredTotal uint) []*gitdomain.Commit {
	allCommits := append(previous, next...)
	// ensure that we don't return more than what was requested
	if uint(len(allCommits)) > desiredTotal {
		return allCommits[:desiredTotal]
	}
	return allCommits
}

type wrappedCommit struct {
	*gitdomain.Commit
	files []string
}

// FirstEverCommit returns the first commit ever made to the repository.
func (c *clientImplementor) FirstEverCommit(ctx context.Context, repo api.RepoName) (commit *gitdomain.Commit, err error) {
	ctx, _, endObservation := c.operations.firstEverCommit.With(ctx, &err, observation.Args{
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

	result, err := client.FirstEverCommit(ctx, &proto.FirstEverCommitRequest{
		RepoName: string(repo),
	})
	if err != nil {
		return nil, err
	}

	return gitdomain.CommitFromProto(result.GetCommit()), nil
}

type ArchiveFormat string

const (
	// ArchiveFormatZip indicates a zip archive is desired.
	ArchiveFormatZip ArchiveFormat = "zip"

	// ArchiveFormatTar indicates a tar archive is desired.
	ArchiveFormatTar ArchiveFormat = "tar"
)

func ArchiveFormatFromProto(pf proto.ArchiveFormat) ArchiveFormat {
	switch pf {
	case proto.ArchiveFormat_ARCHIVE_FORMAT_ZIP:
		return ArchiveFormatZip
	case proto.ArchiveFormat_ARCHIVE_FORMAT_TAR:
		return ArchiveFormatTar
	default:
		return ""
	}
}

func (f ArchiveFormat) ToProto() proto.ArchiveFormat {
	switch f {
	case ArchiveFormatZip:
		return proto.ArchiveFormat_ARCHIVE_FORMAT_ZIP
	case ArchiveFormatTar:
		return proto.ArchiveFormat_ARCHIVE_FORMAT_TAR
	default:
		return proto.ArchiveFormat_ARCHIVE_FORMAT_UNSPECIFIED
	}
}

func (c *clientImplementor) ArchiveReader(ctx context.Context, repo api.RepoName, options ArchiveOptions) (_ io.ReadCloser, err error) {
	ctx, _, endObservation := c.operations.archiveReader.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append(
			[]attribute.KeyValue{repo.Attr()},
			options.Attrs()...,
		),
	})

	if authz.SubRepoEnabled(c.subRepoPermsChecker) && !actor.FromContext(ctx).IsInternal() {
		if enabled, err := authz.SubRepoEnabledForRepo(ctx, c.subRepoPermsChecker, repo); err != nil {
			return nil, errors.Wrap(err, "sub-repo permissions check")
		} else if enabled {
			return nil, errors.New("archiveReader invoked for a repo with sub-repo permissions")
		}
	}

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, err
	}

	req := options.ToProto(string(repo))

	ctx, cancel := context.WithCancel(ctx)
	cli, err := client.Archive(ctx, req)
	if err != nil {
		cancel()
		endObservation(1, observation.Args{})
		return nil, err
	}

	// We start by reading the first message to early-exit on potential errors,
	// ie. revision not found errors or invalid git command.
	firstMessage, firstErr := cli.Recv()
	if firstErr != nil {
		if errors.HasType[*gitdomain.RevisionNotFoundError](firstErr) {
			cancel()
			err = firstErr
			endObservation(1, observation.Args{})
			return nil, err
		}
	}

	firstRespRead := false
	r := streamio.NewReader(func() ([]byte, error) {
		if !firstRespRead {
			firstRespRead = true
			if firstErr != nil {
				return nil, firstErr
			}
			return firstMessage.GetData(), nil
		}

		m, err := cli.Recv()
		if err != nil {
			return nil, err
		}
		return m.GetData(), nil
	})

	return &archiveReader{
		Reader: r,
		cancel: cancel,
		onClose: func() {
			endObservation(1, observation.Args{})
		},
	}, nil
}

type archiveReader struct {
	io.Reader
	cancel  context.CancelFunc
	onClose func()
}

func (br *archiveReader) Close() error {
	br.cancel()
	br.onClose()
	return nil
}

func maybeIncludeFiles(opt CommitsOptions, checker authz.SubRepoPermissionChecker) CommitsOptions {
	if authz.SubRepoEnabled(checker) {
		// If sub-repo permissions enabled, must fetch files modified w/ commits to determine if user has access to view this commit
		opt.IncludeFiles = true
	}
	return opt
}

// ListRefs returns a list of all refs in the repository.
func (c *clientImplementor) ListRefs(ctx context.Context, repo api.RepoName, opt ListRefsOpts) (_ []gitdomain.Ref, err error) {
	ctx, _, endObservation := c.operations.listRefs.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.Bool("headsOnly", opt.HeadsOnly),
			attribute.Bool("tagsOnly", opt.TagsOnly),
		},
	})
	defer endObservation(1, observation.Args{})

	client, err := c.clientSource.ClientForRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	req := &proto.ListRefsRequest{
		RepoName:  string(repo),
		HeadsOnly: opt.HeadsOnly,
		TagsOnly:  opt.TagsOnly,
	}

	for _, c := range opt.PointsAtCommit {
		req.PointsAtCommit = append(req.PointsAtCommit, string(c))
	}

	if opt.Contains != "" {
		req.ContainsSha = pointers.Ptr(string(opt.Contains))
	}

	cc, err := client.ListRefs(ctx, req)
	if err != nil {
		return nil, err
	}

	refs := make([]gitdomain.Ref, 0)
	for {
		resp, err := cc.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		for _, ref := range resp.GetRefs() {
			refs = append(refs, gitdomain.RefFromProto(ref))
		}
	}

	return refs, nil
}

// rel strips the leading "/" prefix from the path string, effectively turning
// an absolute path into one relative to the root directory. A path that is just
// "/" is treated specially, returning just ".".
//
// The elements in a file path are separated by slash ('/', U+002F) characters,
// regardless of host operating system convention.
func rel(path string) string {
	if path == "/" {
		return "."
	}
	return strings.TrimPrefix(path, "/")
}
