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
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const devNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

type RepositoryComparisonInput struct {
	Base *string
	Head *string
}

func NewRepositoryComparison(ctx context.Context, r *repositoryResolver, args *RepositoryComparisonInput) (*repositoryComparisonResolver, error) {
	var baseRevspec, headRevspec string
	if args.Base == nil {
		baseRevspec = "HEAD"
	} else {
		baseRevspec = *args.Base
	}
	if args.Head == nil {
		headRevspec = "HEAD"
	} else {
		headRevspec = *args.Head
	}

	getCommit := func(ctx context.Context, repo gitserver.Repo, revspec string) (*gitCommitResolver, error) {
		if revspec == devNullSHA {
			return nil, nil
		}

		// Call ResolveRevision to trigger fetches from remote (in case base/head commits don't
		// exist).
		commitID, err := git.ResolveRevision(ctx, repo, nil, revspec, nil)
		if err != nil {
			return nil, err
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
	head, err := getCommit(ctx, *grepo, headRevspec)
	if err != nil {
		return nil, err
	}

	return &repositoryComparisonResolver{
		baseRevspec: baseRevspec,
		headRevspec: headRevspec,
		base:        base,
		head:        head,
		repo:        r,
	}, nil
}

func (r *repositoryResolver) Comparison(ctx context.Context, args *RepositoryComparisonInput) (*repositoryComparisonResolver, error) {
	return NewRepositoryComparison(ctx, r, args)
}

type repositoryComparisonResolver struct {
	baseRevspec, headRevspec string
	base, head               *gitCommitResolver
	repo                     *repositoryResolver
}

func (r *repositoryComparisonResolver) Range() *gitRevisionRange {
	return &gitRevisionRange{
		expr:      r.baseRevspec + "..." + r.headRevspec,
		base:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.baseRevspec, repo: r.repo}},
		head:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.headRevspec, repo: r.repo}},
		mergeBase: nil, // not currently used
	}
}

func (r *repositoryComparisonResolver) Commits(args *struct {
	First *int32
}) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		revisionRange: string(r.baseRevspec) + ".." + string(r.headRevspec),
		first:         args.First,
		repo:          r.repo,
	}
}

func (r *repositoryComparisonResolver) FileDiffs(args *struct {
	First *int32
}) *fileDiffConnectionResolver {
	return &fileDiffConnectionResolver{
		cmp:   r,
		first: args.First,
	}
}

type fileDiffConnectionResolver struct {
	cmp   *repositoryComparisonResolver // {base,head}{,RevSpec} and repo
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
			rangeSpec = string(r.cmp.baseRevspec) + ".." + string(hOid)
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

func (r *fileDiffConnectionResolver) Nodes(ctx context.Context) ([]*fileDiffResolver, error) {
	fileDiffs, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(fileDiffs) > int(*r.first) {
		// Don't return +1 results, which is used to determine if next page exists.
		fileDiffs = fileDiffs[:*r.first]
	}

	resolvers := make([]*fileDiffResolver, len(fileDiffs))
	for i, fileDiff := range fileDiffs {
		resolvers[i] = &fileDiffResolver{
			fileDiff: fileDiff,
			cmp:      r.cmp,
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

func (r *fileDiffConnectionResolver) DiffStat(ctx context.Context) (*diffStat, error) {
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

type fileDiffResolver struct {
	fileDiff *diff.FileDiff
	cmp      *repositoryComparisonResolver // {base,head}{,RevSpec} and repo
}

func (r *fileDiffResolver) OldPath() *string { return diffPathOrNull(r.fileDiff.OrigName) }
func (r *fileDiffResolver) NewPath() *string { return diffPathOrNull(r.fileDiff.NewName) }
func (r *fileDiffResolver) Hunks() []*diffHunk {
	hunks := make([]*diffHunk, len(r.fileDiff.Hunks))
	for i, hunk := range r.fileDiff.Hunks {
		hunks[i] = &diffHunk{hunk: hunk}
	}
	return hunks
}

func (r *fileDiffResolver) Stat() *diffStat {
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
		commit: r.cmp.base,
		path:   r.fileDiff.OrigName,
		stat:   createFileInfo(r.fileDiff.OrigName, false),
	}
}

func (r *fileDiffResolver) NewFile() *gitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.NewName) == nil {
		return nil
	}
	return &gitTreeEntryResolver{
		commit: r.cmp.head,
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
