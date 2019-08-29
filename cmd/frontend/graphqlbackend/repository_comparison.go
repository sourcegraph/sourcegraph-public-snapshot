package graphqlbackend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const devNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// RepositoryComparison implements the RepositoryComparison GraphQL type.
type RepositoryComparison interface {
	BaseRepository() *RepositoryResolver
	HeadRepository() *RepositoryResolver
	Range(context.Context) (GitRevisionRange, error)
	Commits(*graphqlutil.ConnectionArgs) *gitCommitConnectionResolver
	FileDiffs(*graphqlutil.ConnectionArgs) FileDiffConnection
}

type RepositoryComparisonInput struct {
	Base    *string
	BaseOID *string
	Head    *string
	HeadOID *string
}

type resolvedRevspec struct {
	expr     string
	commitID api.CommitID
}

func NewResolvedRevspec(expr string, commitID api.CommitID) resolvedRevspec {
	return resolvedRevspec{expr: expr, commitID: commitID}
}

func NewRepositoryComparison(ctx context.Context, r *RepositoryResolver, args *RepositoryComparisonInput) (RepositoryComparison, error) {
	var baseRevspec, headRevspec resolvedRevspec
	if args.Base == nil {
		baseRevspec = resolvedRevspec{expr: "HEAD"}
	} else {
		baseRevspec = resolvedRevspec{expr: *args.Base}
	}
	if args.BaseOID != nil {
		baseRevspec.commitID = api.CommitID(*args.BaseOID)
	}
	if args.Head == nil {
		headRevspec = resolvedRevspec{expr: "HEAD"}
	} else {
		headRevspec = resolvedRevspec{expr: *args.Head}
	}
	if args.HeadOID != nil {
		headRevspec.commitID = api.CommitID(*args.HeadOID)
	}

	getCommit := func(ctx context.Context, repo gitserver.Repo, revspec resolvedRevspec) (*GitCommitResolver, error) {
		if revspec.expr == devNullSHA || revspec.commitID == devNullSHA {
			return nil, nil
		}

		var commitID api.CommitID
		if revspec.commitID != "" {
			commitID = revspec.commitID
		} else {
			// Call ResolveRevision to trigger fetches from remote (in case base/head commits don't
			// exist).
			var err error
			commitID, err = git.ResolveRevision(ctx, repo, nil, revspec.expr, nil)
			if err != nil {
				return nil, err
			}
		}

		commit, err := git.GetCommit(ctx, repo, nil, commitID)
		if err != nil {
			return nil, err
		}
		return toGitCommitResolver(r, commit), nil
	}

	grepo, err := backend.CachedGitRepo(ctx, r.repo)
	if err != nil {
		return nil, err
	}
	base, err := getCommit(ctx, *grepo, baseRevspec)
	if err != nil {
		return nil, err
	}
	if baseRevspec.commitID == "" {
		baseRevspec.commitID = api.CommitID(base.oid)
	}
	head, err := getCommit(ctx, *grepo, headRevspec)
	if err != nil {
		return nil, err
	}
	if headRevspec.commitID == "" {
		headRevspec.commitID = api.CommitID(head.oid)
	}

	return &RepositoryComparisonResolver{
		baseRevspec: baseRevspec,
		headRevspec: headRevspec,
		base:        base,
		head:        head,
		repo:        r,
	}, nil
}

func (r *RepositoryResolver) Comparison(ctx context.Context, args *RepositoryComparisonInput) (RepositoryComparison, error) {
	return NewRepositoryComparison(ctx, r, args)
}

type RepositoryComparisonResolver struct {
	baseRevspec, headRevspec resolvedRevspec
	base, head               *GitCommitResolver
	repo                     *RepositoryResolver
}

func (r *RepositoryComparisonResolver) BaseRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryComparisonResolver) HeadRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryComparisonResolver) Range(context.Context) (GitRevisionRange, error) {
	return NewGitRevisionRange(r.baseRevspec, r.repo, r.headRevspec, r.repo), nil
}

func NewGitRevisionRange(baseRevspec resolvedRevspec, baseRepo *RepositoryResolver, headRevspec resolvedRevspec, headRepo *RepositoryResolver) GitRevisionRange {
	return &gitRevisionRange{
		expr:      baseRevspec.expr + "..." + headRevspec.expr,
		base:      &gitRevSpec{expr: &gitRevSpecExpr{expr: baseRevspec.expr, oid: GitObjectID(baseRevspec.commitID), repo: baseRepo}},
		head:      &gitRevSpec{expr: &gitRevSpecExpr{expr: headRevspec.expr, oid: GitObjectID(headRevspec.commitID), repo: headRepo}},
		mergeBase: nil, // not currently used
	}
}

func (r *RepositoryComparisonResolver) Commits(args *graphqlutil.ConnectionArgs) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		revisionRange: string(r.baseRevspec.commitID) + ".." + string(r.headRevspec.commitID),
		first:         args.First,
		repo:          r.repo,
	}
}

func (r *RepositoryComparisonResolver) FileDiffs(args *graphqlutil.ConnectionArgs) FileDiffConnection {
	return &fileDiffConnectionResolver{
		cmp:   r,
		first: args.First,
	}
}

// FileDiffConnection implements the FileDiffConnection GraphQL type.
type FileDiffConnection interface {
	Nodes(context.Context) ([]FileDiff, error)
	TotalCount(context.Context) (*int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
	DiffStat(context.Context) (IDiffStat, error)
	RawDiff(context.Context) (string, error)
}

type fileDiffConnectionResolver struct {
	cmp   *RepositoryComparisonResolver // {base,head}{,RevSpec} and repo
	first *int32

	// cache result because it is used by multiple fields
	once        sync.Once
	fileDiffs   []*diff.FileDiff
	hasNextPage bool
	err         error
}

func (r *fileDiffConnectionResolver) compute(ctx context.Context) ([]*diff.FileDiff, error) {
	do := func() ([]*diff.FileDiff, error) {
		var rangeSpec string
		hOid := r.cmp.head.OID()
		if r.cmp.base == nil {
			// Rare case: the base is the empty tree, in which case we need ".." not "..." because the latter only works for commits.
			rangeSpec = string(r.cmp.baseRevspec.commitID) + ".." + string(hOid)
		} else {
			rangeSpec = string(r.cmp.base.OID()) + "..." + string(hOid)
		}
		if strings.HasPrefix(rangeSpec, "-") || strings.HasPrefix(rangeSpec, ".") {
			// This should not be possible since r.head is a SHA returned by ResolveRevision, but be
			// extra careful to avoid letting user input add additional `git diff` command-line
			// flags or refer to a file.
			return nil, fmt.Errorf("invalid diff range argument: %q", rangeSpec)
		}
		cachedRepo, err := backend.CachedGitRepo(ctx, r.cmp.repo.repo)
		if err != nil {
			return nil, err
		}
		rdr, err := git.ExecReader(ctx, *cachedRepo, []string{
			"diff",
			"--find-renames",
			"--find-copies",
			"--full-index",
			"--inter-hunk-context=3",
			"--no-prefix",
			rangeSpec,
			"--",
		})
		if err != nil {
			return nil, err
		}
		defer rdr.Close()

		var fileDiffs []*diff.FileDiff
		if r.first != nil {
			fileDiffs = make([]*diff.FileDiff, 0, int(*r.first)) // preallocate
		}
		dr := diff.NewMultiFileDiffReader(rdr)
		for {
			fileDiff, err := dr.ReadFile()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			fileDiffs = append(fileDiffs, fileDiff)
			if r.first != nil && len(fileDiffs) == int(*r.first) {
				// Check for hasNextPage.
				_, err := dr.ReadFile()
				if err != nil && err != io.EOF {
					return nil, err
				}
				r.hasNextPage = err != io.EOF
				break
			}
		}
		return fileDiffs, nil
	}

	r.once.Do(func() { r.fileDiffs, r.err = do() })
	return r.fileDiffs, r.err
}

func (r *fileDiffConnectionResolver) Nodes(ctx context.Context) ([]FileDiff, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(fileDiffs) > int(*r.first) {
		// Don't return +1 results, which is used to determine if next page exists.
		fileDiffs = fileDiffs[:*r.first]
	}

	resolvers := make([]FileDiff, len(fileDiffs))
	for i, fileDiff := range fileDiffs {
		resolvers[i] = &fileDiffResolver{
			fileDiff: fileDiff,
			base:     r.cmp.base,
			head:     r.cmp.head,
		}
	}
	return resolvers, nil
}

func (r *fileDiffConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first == nil || (len(fileDiffs) > int(*r.first)) {
		n := int32(len(fileDiffs))
		return &n, nil
	}
	return nil, nil // total count is not available
}

func (r *fileDiffConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if _, err := r.compute(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.hasNextPage), nil
}

func (r *fileDiffConnectionResolver) DiffStat(ctx context.Context) (IDiffStat, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var stat diffStat
	for _, fileDiff := range fileDiffs {
		s := fileDiff.Stat()
		stat.added += s.Added
		stat.changed += s.Changed
		stat.deleted += s.Deleted
	}
	return &stat, nil
}

func (r *fileDiffConnectionResolver) RawDiff(ctx context.Context) (string, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return "", err
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

type FileDiff interface {
	OldPath() *string
	NewPath() *string
	Hunks() []*diffHunk
	Stat() IDiffStat
	OldFile() *gitTreeEntryResolver
	NewFile() *gitTreeEntryResolver
	MostRelevantFile() *gitTreeEntryResolver
	InternalID() string

	Raw() *diff.FileDiff
}

func NewFileDiff(fileDiff *diff.FileDiff, base, head *GitCommitResolver) FileDiff {
	return &fileDiffResolver{fileDiff: fileDiff, base: base, head: head}
}

type fileDiffResolver struct {
	fileDiff   *diff.FileDiff
	base, head *GitCommitResolver
}

func (r *fileDiffResolver) Raw() *diff.FileDiff { return r.fileDiff }

func (r *fileDiffResolver) OldPath() *string { return diffPathOrNull(r.fileDiff.OrigName) }
func (r *fileDiffResolver) NewPath() *string { return diffPathOrNull(r.fileDiff.NewName) }
func (r *fileDiffResolver) Hunks() []*diffHunk {
	hunks := make([]*diffHunk, len(r.fileDiff.Hunks))
	for i, hunk := range r.fileDiff.Hunks {
		hunks[i] = &diffHunk{hunk: hunk}
	}
	return hunks
}

func (r *fileDiffResolver) Stat() IDiffStat {
	stat := r.fileDiff.Stat()
	return &diffStat{
		added:   stat.Added,
		changed: stat.Changed,
		deleted: stat.Deleted,
	}
}

func (r *fileDiffResolver) OldFile() *gitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.OrigName) == nil {
		return nil
	}
	return &gitTreeEntryResolver{
		commit: r.base,
		path:   r.fileDiff.OrigName,
		stat:   createFileInfo(r.fileDiff.OrigName, false),
	}
}

func (r *fileDiffResolver) NewFile() *gitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.NewName) == nil {
		return nil
	}
	return &gitTreeEntryResolver{
		commit: r.head,
		path:   r.fileDiff.NewName,
		stat:   createFileInfo(r.fileDiff.NewName, false),
	}
}

func (r *fileDiffResolver) MostRelevantFile() *gitTreeEntryResolver {
	if newFile := r.NewFile(); newFile != nil {
		return newFile
	}
	return r.OldFile()
}

func (r *fileDiffResolver) InternalID() string {
	b := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", len(r.fileDiff.OrigName), r.fileDiff.OrigName, r.fileDiff.NewName)))
	return hex.EncodeToString(b[:])[:32]
}

func diffPathOrNull(path string) *string {
	if path == "/dev/null" || path == "" {
		return nil
	}
	return &path
}

type diffHunk struct {
	hunk *diff.Hunk
}

func (r *diffHunk) OldRange() *diffHunkRange {
	return &diffHunkRange{startLine: r.hunk.OrigStartLine, lines: r.hunk.OrigLines}
}
func (r *diffHunk) OldNoNewlineAt() bool { return r.hunk.OrigNoNewlineAt != 0 }
func (r *diffHunk) NewRange() *diffHunkRange {
	return &diffHunkRange{startLine: r.hunk.NewStartLine, lines: r.hunk.NewLines}
}

func (r *diffHunk) Section() *string {
	if r.hunk.Section == "" {
		return nil
	}
	return &r.hunk.Section
}
func (r *diffHunk) Body() string { return string(r.hunk.Body) }

type diffHunkRange struct {
	startLine int32
	lines     int32
}

func (r *diffHunkRange) StartLine() int32 { return r.startLine }
func (r *diffHunkRange) Lines() int32     { return r.lines }

type diffStat struct{ added, changed, deleted int32 }

func (r *diffStat) Added() int32   { return r.added }
func (r *diffStat) Changed() int32 { return r.changed }
func (r *diffStat) Deleted() int32 { return r.deleted }

type DiffStat struct{ Added_, Changed_, Deleted_ int32 }

func (v DiffStat) Added() int32   { return v.Added_ }
func (v DiffStat) Changed() int32 { return v.Changed_ }
func (v DiffStat) Deleted() int32 { return v.Deleted_ }

type IDiffStat interface {
	Added() int32
	Changed() int32
	Deleted() int32
}
