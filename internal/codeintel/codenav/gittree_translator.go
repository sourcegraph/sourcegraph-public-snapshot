package codenav

import (
	"cmp"
	"context"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type CompactGitTreeTranslator interface {
	// TranslatePosition returns None if the given position is on a line that was removed or modified
	// between from and to
	TranslatePosition(
		ctx context.Context, from api.CommitID, to api.CommitID, path core.RepoRelPath, position scip.Position,
	) (core.Option[scip.Position], error)

	// TranslateRange returns None if its start or end positions are on a line that was removed or modified
	// between from and to
	TranslateRange(
		ctx context.Context, from api.CommitID, to api.CommitID, path core.RepoRelPath, range_ scip.Range,
	) (core.Option[scip.Range], error)

	// Prefetch will set-up the cache and kick off a diff command for the given paths. It returns immediately
	// and does not wait for the diff to complete.
	Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath)
}

func NewCompactGitTreeTranslator(client gitserver.Client, repo sgtypes.Repo) CompactGitTreeTranslator {
	return &newTranslator{
		client:    client,
		repo:      repo,
		hunkCache: make(map[string]func() ([]compactHunk, error)),
	}
}

type newTranslator struct {
	client    gitserver.Client
	repo      sgtypes.Repo
	cacheLock sync.RWMutex
	hunkCache map[string]func() ([]compactHunk, error)
}

func (t *newTranslator) TranslatePosition(
	ctx context.Context, from api.CommitID, to api.CommitID, path core.RepoRelPath, pos scip.Position,
) (core.Option[scip.Position], error) {
	if from == to {
		return core.Some(pos), nil
	}
	hunks, err := t.readCachedHunks(ctx, from, to, path)
	if err != nil {
		return core.None[scip.Position](), err
	}
	return translatePosition(hunks, pos), nil
}

func (t *newTranslator) TranslateRange(
	ctx context.Context, from api.CommitID, to api.CommitID, path core.RepoRelPath, range_ scip.Range,
) (core.Option[scip.Range], error) {
	if from == to {
		return core.Some(range_), nil
	}
	hunks, err := t.readCachedHunks(ctx, from, to, path)
	if err != nil {
		return core.None[scip.Range](), err
	}
	return translateRange(hunks, range_), nil
}

func (t *newTranslator) readCachedHunks(
	ctx context.Context, from api.CommitID, to api.CommitID, path core.RepoRelPath,
) (_ []compactHunk, err error) {
	_ = t.fetchHunks(ctx, from, to, path)
	t.cacheLock.RLock()
	hunkFunc := t.hunkCache[makeTypedKey(from, to, path)]
	t.cacheLock.RUnlock()
	return hunkFunc()
}

func (t *newTranslator) Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath) {
	run := t.fetchHunks(ctx, from, to, paths...)
	// Kick off the actual diff command
	go func() { run() }()
}

func (t *newTranslator) fetchHunks(ctx context.Context, from api.CommitID, to api.CommitID, paths ...core.RepoRelPath) func() {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()
	paths = genslices.Filter(paths, func(path core.RepoRelPath) bool {
		_, ok := t.hunkCache[makeTypedKey(from, to, path)]
		return !ok
	})
	if len(paths) == 0 {
		return func() {}
	}
	onceHunks := sync.OnceValues(func() (map[core.RepoRelPath][]compactHunk, error) {
		return t.runDiff(ctx, from, to, paths)
	})
	for _, path := range paths {
		key := makeTypedKey(from, to, path)
		t.hunkCache[key] = sync.OnceValues(func() ([]compactHunk, error) {
			hunkss, err := onceHunks()
			if err != nil {
				return []compactHunk{}, nil
			}
			hunks, ok := hunkss[path]
			if !ok {
				return []compactHunk{}, nil
			}
			return hunks, nil
		})
	}
	return func() {
		_, _ = onceHunks()
	}
}

func (t *newTranslator) runDiff(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath) (map[core.RepoRelPath][]compactHunk, error) {
	r, err := t.client.Diff(ctx, t.repo.Name, gitserver.DiffOptions{
		Base:             string(from),
		Head:             string(to),
		Paths:            genslices.Map(paths, func(p core.RepoRelPath) string { return p.RawValue() }),
		RangeType:        "..",
		InterHunkContext: pointers.Ptr(0),
		ContextLines:     pointers.Ptr(0),
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := r.Close()
		if err == nil {
			err = closeErr
		}
	}()
	fds := make(map[core.RepoRelPath][]compactHunk)
	for {
		fd, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fds, nil
			} else {
				return nil, err
			}
		}
		if fd.OrigName != fd.NewName {
			// We cannot handle file renames
			continue
		}
		fds[core.NewRepoRelPathUnchecked(fd.OrigName)] = genslices.Map(fd.Hunks, newCompactHunk)
	}
}

func precedingHunk(hunks []compactHunk, line int32) core.Option[compactHunk] {
	line += 1 // diff hunks are 1-based, compared to our 0-based scip ranges
	precedingHunkIx, found := slices.BinarySearchFunc(hunks, line, func(h compactHunk, l int32) int {
		return cmp.Compare(h.origStartLine, l)
	})
	if precedingHunkIx == 0 && !found {
		// No preceding hunk means the position was not affected by any hunks
		return core.None[compactHunk]()
	}
	ix := precedingHunkIx
	if !found {
		ix -= 1
	}
	return core.Some(hunks[ix])
}

func newTranslateLine(
	hunks []compactHunk,
	line int32,
) core.Option[int32] {
	hunk, ok := precedingHunk(hunks, line).Get()
	if !ok {
		return core.Some(line)
	}
	return hunk.ShiftLine(line)
}

func translatePosition(
	hunks []compactHunk,
	pos scip.Position,
) core.Option[scip.Position] {
	hunk, ok := precedingHunk(hunks, pos.Line).Get()
	if !ok {
		return core.Some(pos)
	}
	return hunk.ShiftPosition(pos)
}

func translateRange(
	hunks []compactHunk,
	range_ scip.Range,
) core.Option[scip.Range] {
	// Fast path for single-line ranges
	if range_.Start.Line == range_.End.Line {
		newLine, ok := newTranslateLine(hunks, range_.Start.Line).Get()
		if !ok {
			return core.None[scip.Range]()
		}
		return core.Some(scip.Range{
			Start: scip.Position{Line: newLine, Character: range_.Start.Character},
			End:   scip.Position{Line: newLine, Character: range_.End.Character},
		})
	}

	start, ok := translatePosition(hunks, range_.Start).Get()
	if !ok {
		return core.None[scip.Range]()
	}
	end, ok := translatePosition(hunks, range_.End).Get()
	if !ok {
		return core.None[scip.Range]()
	}
	return core.Some(scip.Range{Start: start, End: end})
}

type compactHunk struct {
	// starting line number in original file
	origStartLine int32
	// number of lines the hunk applies to in the original file
	origLines int32
	// starting line number in new file
	newStartLine int32
	// number of lines the hunk applies to in the new file
	newLines int32
}

func newCompactHunk(h *diff.Hunk) compactHunk {
	// If either origLines or newLines are 0, their corresponding line is shifted by an additional -1
	// in the `git diff` output, to make it clear to the user that the line is not included in the
	// displayed hunk.
	// For our purposes we need the actual start line of the hunk though
	origStartLine := h.OrigStartLine
	if h.OrigLines == 0 {
		origStartLine += 1
	}
	newStartLine := h.NewStartLine
	if h.NewLines == 0 {
		newStartLine += 1
	}
	return compactHunk{
		origStartLine: origStartLine,
		origLines:     h.OrigLines,
		newStartLine:  newStartLine,
		newLines:      h.NewLines,
	}
}

func (h *compactHunk) OverlapsLine(line int32) bool {
	// git diff hunks are 1-based, vs our 0-based scip ranges
	return h.origStartLine <= line+1 && h.origStartLine+h.origLines > line+1
}

func (h *compactHunk) ShiftLine(line int32) core.Option[int32] {
	if h.OverlapsLine(line) {
		return core.None[int32]()
	}
	originalSpan := h.origStartLine + h.origLines
	newSpan := h.newStartLine + h.newLines
	return core.Some(line + newSpan - originalSpan)
}

func (h *compactHunk) ShiftPosition(position scip.Position) core.Option[scip.Position] {
	newLine, ok := h.ShiftLine(position.Line).Get()
	if !ok {
		return core.None[scip.Position]()
	}
	if newLine == position.Line {
		return core.Some(position)
	}
	return core.Some(scip.Position{Line: newLine, Character: position.Character})
}

// GitTreeTranslator translates a position within a git tree at a source commit into the
// equivalent position in a target commit. The git tree translator instance carries
// along with it the source commit.
//
// NOTE(id: codenav-file-rename-detection) At the moment, this code cannot handle positions/ranges
// going from one file to another (notice that the return values don't contain any updated
// path), because there is no way in gitserver to get rename detection without requesting
// a diff of the full repo (which may be quite large).
//
// Additionally, it's not clear if reusing the Document at path P1 in an older commit,
// at a different path P2 in a newer commit is even reliably useful, since symbol names
// may change based on the file name or directory name depending on the language.
//
// TODO(id: GitTreeTranslator-cleanup): Instead of storing the TranslationBase, we should
// take that as an argument. Specifically, use a struct with two fields, AncestorCommit
// and DescendantCommit, and avoid Source/Target terminology (which becomes confusing to
// understand with the reverse parameter). Instead, we can use an enum MappingDirection
// FromDescendantToAncestor | FromAncestorToDescendant if really needed (to avoid
// inconsistency when modifying the APIs below, as they take different values for 'reverse'
// in production).
type GitTreeTranslator interface {
	Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath)
	// GetTargetCommitPositionFromSourcePosition translates the given position from the source commit into the given
	// target commit. The target commit's position is returned, along with a boolean flag
	// indicating that the translation was successful. If reverse is true, then the source and
	// target commits are swapped.
	//
	// TODO(id: GitTreeTranslator-cleanup): The reverse parameter is always false in production,
	// let's remove the extra parameter.
	GetTargetCommitPositionFromSourcePosition(ctx context.Context, commit string, path string, px shared.Position, reverse bool) (shared.Position, bool, error)

	// GetTargetCommitRangeFromSourceRange translates the given range from the source commit into the given target
	// commit. The target commit's range is returned, along with a boolean flag indicating
	// that the translation was successful. If reverse is true, then the source and target commits
	// are swapped.
	//
	// TODO(id: GitTreeTranslator-cleanup): The reverse parameter is always true in production,
	// let's remove the extra parameter.
	GetTargetCommitRangeFromSourceRange(ctx context.Context, commit string, path string, rx shared.Range, reverse bool) (shared.Range, bool, error)

	GetSourceCommit() api.CommitID

	// TODO(id: add-bulk-translation-api) Add an API which can map a bunch of ranges all at once
	// so as to avoid iterating over the hunks repeatedly. So long as there is no error getting
	// the hunks, the API should try to convert as many ranges as possible instead of fail-fast
	// behavior. It is OK to expect the input set of ranges to be sorted.
	// Might be useful to add a simple benchmark too.
}

type gitTreeTranslator struct {
	compact   CompactGitTreeTranslator
	client    gitserver.Client
	base      *TranslationBase
	hunkCache HunkCache
}

// TODO(id: GitTreeTranslator-cleanup): Strictly speaking, calling this TranslationBase is not
// quite correct as things can flip around based on the reverse parameter. So get rid
// of the commit field and pass that as a parameter for increased clarity at call-sites.
type TranslationBase struct {
	Repo   *sgtypes.Repo
	Commit api.CommitID
}

func (r *TranslationBase) GetRepoID() int {
	return int(r.Repo.ID)
}

// HunkCache is a concurrency-safe LRU cache that holds git diff hunks.
//
// WARNING: It is NOT safe to modify the return value of Get or to
// modify key or value passed to Set. Not 100% sure about this, filed:
// https://github.com/dgraph-io/ristretto/issues/381
type HunkCache interface {
	// Get returns the value (if any) and a boolean representing whether the value was
	// found or not.
	Get(key any) (any, bool)

	// Set attempts to add the key-value item to the cache with the given cost. If it
	// returns false, then the value as dropped and the item isn't added to the cache.
	Set(key, value any, cost int64) bool
}

// NewHunkCache creates a data cache instance with the given maximum capacity.
func NewHunkCache(size int) (HunkCache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(size) * 10,
		MaxCost:     int64(size),
		BufferItems: 64,
	})
}

// NewGitTreeTranslator creates a new GitTreeTranslator with the given repository and source commit.
func NewGitTreeTranslator(client gitserver.Client, base *TranslationBase, hunkCache HunkCache) GitTreeTranslator {
	return &gitTreeTranslator{
		client:    client,
		compact:   NewCompactGitTreeTranslator(client, *base.Repo),
		hunkCache: hunkCache,
		base:      base,
	}
}

// GetTargetCommitPositionFromSourcePosition translates the given position from the source commit into the given
// target commit. The target commit position is returned, along with a boolean flag
// indicating that the translation was successful. If reverse is true, then the source and
// target commits are swapped.
func (g *gitTreeTranslator) GetTargetCommitPositionFromSourcePosition(ctx context.Context, commit string, path string, px shared.Position, reverse bool) (shared.Position, bool, error) {
	from, to := g.base.Commit, api.CommitID(commit)
	if reverse {
		from, to = to, from
	}
	posOpt, err := g.compact.TranslatePosition(ctx, from, to, core.NewRepoRelPathUnchecked(path), px.ToSCIPPosition())
	if err != nil {
		return shared.Position{}, false, err
	}
	pos, ok := posOpt.Get()
	return shared.TranslatePosition(pos), ok, nil
}

// GetTargetCommitRangeFromSourceRange translates the given range from the source commit into the given target
// commit. The target commit range is returned, along with a boolean flag indicating
// that the translation was successful. If reverse is true, then the source and target commits
// are swapped.
func (g *gitTreeTranslator) GetTargetCommitRangeFromSourceRange(ctx context.Context, commit string, path string, rx shared.Range, reverse bool) (shared.Range, bool, error) {
	from, to := g.base.Commit, api.CommitID(commit)
	if reverse {
		from, to = to, from
	}
	posOpt, err := g.compact.TranslateRange(ctx, from, to, core.NewRepoRelPathUnchecked(path), rx.ToSCIPRange())
	if err != nil {
		return shared.Range{}, false, err
	}
	range_, ok := posOpt.Get()
	return shared.TranslateRange(range_), ok, nil
}

func (g *gitTreeTranslator) GetSourceCommit() api.CommitID {
	return g.base.Commit
}

func (g *gitTreeTranslator) Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath) {
	g.compact.Prefetch(ctx, from, to, paths)
}

func makeTypedKey(from api.CommitID, to api.CommitID, path core.RepoRelPath) string {
	return makeKey(string(from), string(to), path.RawValue())
}
func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}
