package graphqlbackend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const devNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

type RepositoryComparisonConnectionResolver interface {
	Nodes(ctx context.Context) ([]*RepositoryComparisonResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type RepositoryComparisonInput struct {
	Base *string
	Head *string
}

func NewRepositoryComparison(ctx context.Context, r *RepositoryResolver, args *RepositoryComparisonInput) (*RepositoryComparisonResolver, error) {
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

	getCommit := func(ctx context.Context, repo gitserver.Repo, revspec string) (*GitCommitResolver, error) {
		if revspec == devNullSHA {
			return nil, nil
		}

		// Optimistically fetch using revspec
		commit, err := git.GetCommit(ctx, repo, nil, api.CommitID(revspec))
		if err == nil {
			return toGitCommitResolver(r, commit), nil
		}

		// Call ResolveRevision to trigger fetches from remote (in case base/head commits don't
		// exist).
		commitID, err := git.ResolveRevision(ctx, repo, nil, revspec, nil)
		if err != nil {
			return nil, err
		}

		commit, err = git.GetCommit(ctx, repo, nil, commitID)
		if err != nil {
			return nil, err
		}
		return toGitCommitResolver(r, commit), nil
	}

	grepo, err := backend.CachedGitRepo(ctx, r.repo)
	if err != nil {
		return nil, err
	}

	var (
		wg               sync.WaitGroup
		base, head       *GitCommitResolver
		baseErr, headErr error
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		base, baseErr = getCommit(ctx, *grepo, baseRevspec)
	}()
	go func() {
		defer wg.Done()
		head, headErr = getCommit(ctx, *grepo, headRevspec)
	}()
	wg.Wait()
	if baseErr != nil {
		return nil, baseErr
	}
	if headErr != nil {
		return nil, headErr
	}

	return &RepositoryComparisonResolver{
		baseRevspec: baseRevspec,
		headRevspec: headRevspec,
		base:        base,
		head:        head,
		repo:        r,
	}, nil
}

func (r *RepositoryResolver) Comparison(ctx context.Context, args *RepositoryComparisonInput) (*RepositoryComparisonResolver, error) {
	return NewRepositoryComparison(ctx, r, args)
}

type RepositoryComparisonResolver struct {
	baseRevspec, headRevspec string
	base, head               *GitCommitResolver
	repo                     *RepositoryResolver
}

func (r *RepositoryComparisonResolver) BaseRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryComparisonResolver) HeadRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryComparisonResolver) Range() *gitRevisionRange {
	return &gitRevisionRange{
		expr:      r.baseRevspec + "..." + r.headRevspec,
		base:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.baseRevspec, repo: r.repo}},
		head:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.headRevspec, repo: r.repo}},
		mergeBase: nil, // not currently used
	}
}

func (r *RepositoryComparisonResolver) Commits(
	args *graphqlutil.ConnectionArgs,
) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		revisionRange: string(r.baseRevspec) + ".." + string(r.headRevspec),
		first:         args.First,
		repo:          r.repo,
	}
}

func (r *RepositoryComparisonResolver) FileDiffs(
	args *FileDiffsConnectionArgs,
) *fileDiffConnectionResolver {
	return &fileDiffConnectionResolver{
		cmp:   r,
		first: args.First,
		after: args.After,
	}
}

type fileDiffConnectionResolver struct {
	cmp   *RepositoryComparisonResolver // {base,head}{,RevSpec} and repo
	first *int32
	after *string

	// cache result because it is used by multiple fields
	once        sync.Once
	fileDiffs   []*diff.FileDiff
	afterIdx    int32
	hasNextPage bool
	err         error
}

// compute returns the computed FileDiffs, the index from which to return entries (`after` param) and an optional error.
func (r *fileDiffConnectionResolver) compute(ctx context.Context) ([]*diff.FileDiff, int32, error) {
	do := func() ([]*diff.FileDiff, int32, error) {
		var afterIdx int32
		// Todo: It's possible that the rangeSpec changes in between two executions, then the cursor would be invalid and the
		// whole pagination should not be continued.
		if r.after != nil {
			parsedIdx, err := strconv.ParseInt(*r.after, 0, 32)
			if err != nil {
				return nil, afterIdx, err
			}
			if parsedIdx < 0 {
				parsedIdx = 0
			}
			afterIdx = int32(parsedIdx)
		}

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
			return nil, afterIdx, fmt.Errorf("invalid diff range argument: %q", rangeSpec)
		}
		cachedRepo, err := backend.CachedGitRepo(ctx, r.cmp.repo.repo)
		if err != nil {
			return nil, afterIdx, err
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
			return nil, afterIdx, err
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
				return nil, afterIdx, err
			}
			fileDiffs = append(fileDiffs, fileDiff)
			if r.first != nil && len(fileDiffs) == int(*r.first+afterIdx) {
				// Check for hasNextPage.
				_, err := dr.ReadFile()
				if err != nil && err != io.EOF {
					return nil, afterIdx, err
				}
				r.hasNextPage = err != io.EOF
				break
			}
		}
		return fileDiffs, afterIdx, nil
	}

	r.once.Do(func() { r.fileDiffs, r.afterIdx, r.err = do() })
	return r.fileDiffs, r.afterIdx, r.err
}

func (r *fileDiffConnectionResolver) Nodes(ctx context.Context) ([]*fileDiffResolver, error) {
	fileDiffs, afterIdx, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first != nil && int(*r.first+afterIdx) <= len(fileDiffs) {
		// If first is set, we have an upper bound. If the requested amount of diffs exist, return a slice of `first` entries from afterIdx.
		fileDiffs = fileDiffs[afterIdx:(*r.first + afterIdx)]
	} else if int(afterIdx) <= len(fileDiffs) {
		// If no upper boundary is given, return from the lower boundary.
		fileDiffs = fileDiffs[afterIdx:]
	} else {
		// If the lower boundary is out of bounds, return an empty result.
		fileDiffs = []*diff.FileDiff{}
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
	fileDiffs, _, err := r.compute(ctx)
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
	_, afterIdx, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if !r.hasNextPage {
		return graphqlutil.HasNextPage(r.hasNextPage), nil
	}
	next := int32(afterIdx)
	if r.first != nil {
		next += *r.first
	}
	return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
}

func (r *fileDiffConnectionResolver) DiffStat(ctx context.Context) (*DiffStat, error) {
	fileDiffs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stat := &DiffStat{}
	for _, fileDiff := range fileDiffs {
		s := fileDiff.Stat()
		stat.AddStat(s)
	}
	return stat, nil
}

func (r *fileDiffConnectionResolver) RawDiff(ctx context.Context) (string, error) {
	fileDiffs, _, err := r.compute(ctx)
	if err != nil {
		return "", err
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

type fileDiffResolver struct {
	fileDiff *diff.FileDiff
	cmp      *RepositoryComparisonResolver // {base,head}{,RevSpec} and repo
}

func (r *fileDiffResolver) OldPath() *string { return diffPathOrNull(r.fileDiff.OrigName) }
func (r *fileDiffResolver) NewPath() *string { return diffPathOrNull(r.fileDiff.NewName) }
func (r *fileDiffResolver) Hunks() []*DiffHunk {
	hunks := make([]*DiffHunk, len(r.fileDiff.Hunks))
	highlighter := &fileDiffHighlighter{
		fileDiffResolver: r,
	}
	for i, hunk := range r.fileDiff.Hunks {
		hunks[i] = NewDiffHunk(hunk, highlighter)
	}
	return hunks
}

func (r *fileDiffResolver) Stat() *DiffStat {
	stat := r.fileDiff.Stat()
	return NewDiffStat(stat)
}

func (r *fileDiffResolver) OldFile() *GitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.OrigName) == nil {
		return nil
	}
	return &GitTreeEntryResolver{
		commit: r.cmp.base,
		stat:   CreateFileInfo(r.fileDiff.OrigName, false),
	}
}

func (r *fileDiffResolver) NewFile() *GitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.NewName) == nil {
		return nil
	}
	return &GitTreeEntryResolver{
		commit: r.cmp.head,
		stat:   CreateFileInfo(r.fileDiff.NewName, false),
	}
}

func (r *fileDiffResolver) MostRelevantFile() *GitTreeEntryResolver {
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

func NewDiffHunk(hunk *diff.Hunk, highlighter DiffHighlighter) *DiffHunk {
	return &DiffHunk{hunk: hunk, highlighter: highlighter}
}

type fileDiffHighlighter struct {
	fileDiffResolver *fileDiffResolver
	highlightedBase  []string
	highlightedHead  []string
	highlightOnce    sync.Once
	highlightErr     error
	highlightAborted bool
}

func (r *fileDiffHighlighter) Highlight(ctx context.Context, args *HighlightArgs) ([]string, []string, bool, error) {
	r.highlightOnce.Do(func() {
		highlightFile := func(ctx context.Context, file *GitTreeEntryResolver) ([]string, error) {
			if file == nil {
				return nil, nil
			}
			binary, err := file.Binary(ctx)
			if err != nil {
				return nil, err
			}
			if binary {
				return nil, nil
			}
			highlightedFile, err := file.Highlight(ctx, &HighlightArgs{
				DisableTimeout:     args.DisableTimeout,
				HighlightLongLines: args.HighlightLongLines,
				IsLightTheme:       args.IsLightTheme,
			})
			if err != nil {
				return nil, err
			}
			if highlightedFile.Aborted() {
				r.highlightAborted = true
			}
			return highlight.ParseLinesFromHighlight(highlightedFile.HTML())
		}
		r.highlightedBase, r.highlightErr = highlightFile(ctx, r.fileDiffResolver.OldFile())
		if r.highlightErr != nil {
			return
		}
		r.highlightedHead, r.highlightErr = highlightFile(ctx, r.fileDiffResolver.NewFile())
	})
	return r.highlightedBase, r.highlightedHead, r.highlightAborted, r.highlightErr
}

type DiffHunk struct {
	hunk        *diff.Hunk
	highlighter DiffHighlighter
}

func (r *DiffHunk) OldRange() *DiffHunkRange {
	return NewDiffHunkRange(r.hunk.OrigStartLine, r.hunk.OrigLines)
}
func (r *DiffHunk) OldNoNewlineAt() bool { return r.hunk.OrigNoNewlineAt != 0 }
func (r *DiffHunk) NewRange() *DiffHunkRange {
	return NewDiffHunkRange(r.hunk.NewStartLine, r.hunk.NewLines)
}

func (r *DiffHunk) Section() *string {
	if r.hunk.Section == "" {
		return nil
	}
	return &r.hunk.Section
}

func (r *DiffHunk) Body() string { return string(r.hunk.Body) }

func (r *DiffHunk) Highlight(ctx context.Context, args *HighlightArgs) (*highlightedDiffHunkBodyResolver, error) {
	highlightedBase, highlightedHead, aborted, err := r.highlighter.Highlight(ctx, args)
	if err != nil {
		return nil, err
	}
	hunkLines := strings.Split(string(r.hunk.Body), "\n")
	// Remove final empty line on files that end with a newline, as most code hosts do.
	if hunkLines[len(hunkLines)-1] == "" {
		hunkLines = hunkLines[:len(hunkLines)-1]
	}
	highlightedDiffHunkLineResolvers := make([]*highlightedDiffHunkLineResolver, len(hunkLines))
	// Lines in highlightedBase and highlightedHead are 0-indexed.
	baseLine := r.hunk.OrigStartLine - 1
	headLine := r.hunk.NewStartLine - 1
	for i, hunkLine := range hunkLines {
		highlightedDiffHunkLineResolver := highlightedDiffHunkLineResolver{}
		if len(hunkLine) == 0 || hunkLine[0] == ' ' {
			highlightedDiffHunkLineResolver.kind = "UNCHANGED"
			highlightedDiffHunkLineResolver.html = highlightedBase[baseLine]
			baseLine++
			headLine++
		} else if hunkLine[0] == '+' {
			highlightedDiffHunkLineResolver.kind = "ADDED"
			highlightedDiffHunkLineResolver.html = highlightedHead[headLine]
			headLine++
		} else if hunkLine[0] == '-' {
			highlightedDiffHunkLineResolver.kind = "DELETED"
			highlightedDiffHunkLineResolver.html = highlightedBase[baseLine]
			baseLine++
		} else {
			return nil, fmt.Errorf("expected patch lines to start with ' ', '-', '+', but found %q", hunkLine[0])
		}

		highlightedDiffHunkLineResolvers[i] = &highlightedDiffHunkLineResolver
	}
	return &highlightedDiffHunkBodyResolver{
		highlightedDiffHunkLineResolvers: highlightedDiffHunkLineResolvers,
		aborted:                          aborted,
	}, nil
}

type highlightedDiffHunkBodyResolver struct {
	highlightedDiffHunkLineResolvers []*highlightedDiffHunkLineResolver
	aborted                          bool
}

func (r *highlightedDiffHunkBodyResolver) Aborted() bool {
	return r.aborted
}

func (r *highlightedDiffHunkBodyResolver) Lines() []*highlightedDiffHunkLineResolver {
	return r.highlightedDiffHunkLineResolvers
}

type highlightedDiffHunkLineResolver struct {
	html string
	kind string
}

func (r *highlightedDiffHunkLineResolver) HTML() string {
	return r.html
}

func (r *highlightedDiffHunkLineResolver) Kind() string {
	return r.kind
}

func NewDiffHunkRange(startLine, lines int32) *DiffHunkRange {
	return &DiffHunkRange{startLine: startLine, lines: lines}
}

type DiffHunkRange struct {
	startLine int32
	lines     int32
}

func (r *DiffHunkRange) StartLine() int32 { return r.startLine }
func (r *DiffHunkRange) Lines() int32     { return r.lines }

func NewDiffStat(s diff.Stat) *DiffStat {
	return &DiffStat{
		added:   s.Added,
		changed: s.Changed,
		deleted: s.Deleted,
	}
}

type DiffStat struct{ added, changed, deleted int32 }

func (r *DiffStat) AddStat(s diff.Stat) {
	r.added += s.Added
	r.changed += s.Changed
	r.deleted += s.Deleted
}

func (r *DiffStat) AddDiffStat(s *DiffStat) {
	r.added += s.Added()
	r.changed += s.Changed()
	r.deleted += s.Deleted()
}

func (r *DiffStat) Added() int32   { return r.added }
func (r *DiffStat) Changed() int32 { return r.changed }
func (r *DiffStat) Deleted() int32 { return r.deleted }
