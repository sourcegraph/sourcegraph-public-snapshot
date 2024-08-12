package codenav

import (
	"cmp"
	"context"
	"io"
	"slices"
	"sync"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// GitTreeTranslator translates positions within a git tree between commits.
type GitTreeTranslator interface {
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

	// Prefetch populates the cache with hunks for the given paths. It does not block
	Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath)
}

func NewGitTreeTranslator(client minimalGitserver, repo sgtypes.Repo) GitTreeTranslator {
	return &newTranslator{
		client:    client,
		repo:      repo,
		hunkCache: make(map[hunkCacheKey]func() ([]compactHunk, error)),
	}
}

type hunkCacheKey struct {
	from api.CommitID
	to   api.CommitID
	path core.RepoRelPath
}

type newTranslator struct {
	client    minimalGitserver
	repo      sgtypes.Repo
	cacheLock sync.RWMutex
	hunkCache map[hunkCacheKey]func() ([]compactHunk, error)
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
	_ = t.fetchHunksLazy(ctx, from, to, path)
	t.cacheLock.RLock()
	hunkFunc, ok := t.hunkCache[hunkCacheKey{from, to, path}]
	t.cacheLock.RUnlock()
	if !ok {
		// This should not happen
		return nil, errors.New("no cached hunks for this path")
	}
	return hunkFunc()
}

func (t *newTranslator) Prefetch(ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath) {
	// Kick off the actual diff command in the background
	go t.fetchHunksLazy(ctx, from, to, paths...)()
}

// fetchHunksLazy fetches the hunks for the given paths from the given commit range as a batch from gitserver and
// populates the cache with its results.
// It returns a function that can be called to kick off the actual diff command, otherwise the diff command
// will be kicked off when the first path from paths is requested.
func (t *newTranslator) fetchHunksLazy(ctx context.Context, from api.CommitID, to api.CommitID, paths ...core.RepoRelPath) func() {
	t.cacheLock.Lock()
	defer t.cacheLock.Unlock()
	// Don't fetch diffs for paths we've already cached
	paths = genslices.Filter(paths, func(path core.RepoRelPath) bool {
		_, ok := t.hunkCache[hunkCacheKey{from, to, path}]
		return !ok
	})
	if len(paths) == 0 {
		return func() {}
	}
	onceHunksMap := sync.OnceValues(func() (map[core.RepoRelPath][]compactHunk, error) {
		return t.runDiff(ctx, from, to, paths)
	})
	for _, path := range paths {
		key := hunkCacheKey{from, to, path}
		t.hunkCache[key] = sync.OnceValues(func() ([]compactHunk, error) {
			hunksMap, err := onceHunksMap()
			if err != nil {
				return []compactHunk{}, err
			}
			hunks, ok := hunksMap[path]
			if !ok {
				return []compactHunk{}, nil
			}
			return hunks, nil
		})
	}
	return func() {
		_, _ = onceHunksMap()
	}
}

func (t *newTranslator) runDiff(
	ctx context.Context, from api.CommitID, to api.CommitID, paths []core.RepoRelPath,
) (map[core.RepoRelPath][]compactHunk, error) {
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
	fileDiffs := make(map[core.RepoRelPath][]compactHunk)
	for {
		fileDiff, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fileDiffs, nil
			} else {
				return nil, err
			}
		}
		if fileDiff.OrigName != fileDiff.NewName {
			// We do not handle file renames
			continue
		}
		fileDiffs[core.NewRepoRelPathUnchecked(fileDiff.OrigName)] = genslices.Map(fileDiff.Hunks, newCompactHunk)
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
	if !found {
		precedingHunkIx -= 1
	}
	return core.Some(hunks[precedingHunkIx])
}

func translateLine(hunks []compactHunk, line int32) core.Option[int32] {
	hunk, ok := precedingHunk(hunks, line).Get()
	if !ok {
		return core.Some(line)
	}
	return hunk.shiftLine(line)
}

func translatePosition(hunks []compactHunk, pos scip.Position) core.Option[scip.Position] {
	hunk, ok := precedingHunk(hunks, pos.Line).Get()
	if !ok {
		return core.Some(pos)
	}
	return hunk.shiftPosition(pos)
}

func translateRange(hunks []compactHunk, range_ scip.Range) core.Option[scip.Range] {
	// Fast path for single-line ranges
	if range_.Start.Line == range_.End.Line {
		newLine, ok := translateLine(hunks, range_.Start.Line).Get()
		if !ok {
			return core.None[scip.Range]()
		}
		return core.Some(scip.Range{
			Start: scip.Position{Line: newLine, Character: range_.Start.Character},
			End:   scip.Position{Line: newLine, Character: range_.End.Character},
		})
	}

	if start, ok := translatePosition(hunks, range_.Start).Get(); ok {
		if end, ok := translatePosition(hunks, range_.End).Get(); ok {
			return core.Some(scip.Range{Start: start, End: end})
		}
	}
	return core.None[scip.Range]()
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
	//
	// Examples:
	//
	// $ echo "line1\nline2\nline3\n" > test.txt && echo "line1\nline3\n" > test1.txt && git diff --no-index --no-prefix -U0 test.txt test1.txt
	// diff --git test.txt test1.txt
	// index be4bc321656..cc8972178a0 100644
	// --- test.txt
	// +++ test1.txt
	// @@ -2 +1,0 @@ line1
	// -line2
	//
	// origStartLine: 2, origLines: 1, newStartLine: 1, newLines: 0
	// Would lead to `(1 + 0) - (2 + 1) = -2` even though the hunk only removes one line
	//
	// $ echo "line1\nline2\nline3\n" > test.txt && echo "line1\nline2\nlineNew\nline3\n" > test1.txt && git diff --no-index --no-prefix -U0 test.txt test1.txt
	// diff --git test.txt test1.txt
	// index be4bc321656..8298ca98d51 100644
	// --- test.txt
	// +++ test1.txt
	// @@ -2,0 +3 @@ line2
	// +lineNew
	//
	// origStartLine: 2, origLines: 0, newStartLine: 3, newLines: 1
	// Would lead to `(3 + 1) - (2 + 0) = 2` even though the hunk only adds one line
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

func (h *compactHunk) overlapsLine(line int32) bool {
	// git diff hunks are 1-based, vs our 0-based scip ranges
	return h.origStartLine <= line+1 && line+1 < h.origStartLine+h.origLines
}

func (h *compactHunk) shiftLine(line int32) core.Option[int32] {
	if h.overlapsLine(line) {
		return core.None[int32]()
	}
	originalSpan := h.origStartLine + h.origLines
	newSpan := h.newStartLine + h.newLines
	if newLine := line + newSpan - originalSpan; newLine >= 0 {
		return core.Some(newLine)
	} else {
		return core.None[int32]()
	}
}

func (h *compactHunk) shiftPosition(position scip.Position) core.Option[scip.Position] {
	newLine, ok := h.shiftLine(position.Line).Get()
	if !ok {
		return core.None[scip.Position]()
	}
	return core.Some(scip.Position{Line: newLine, Character: position.Character})
}
