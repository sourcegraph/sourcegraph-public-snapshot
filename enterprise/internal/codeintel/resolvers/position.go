package resolvers

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type positionAdjuster struct {
	hunks []*diff.Hunk
}

// newPositionAdjuster creates a positionAdjuster by performing a git diff operation on the give
// path between the source and target commits. If the commits are the same or there are no changes
// in the given path, then a trivial-case no-op position adjuster is returned.
func newPositionAdjuster(ctx context.Context, repo *types.Repo, sourceCommit, targetCommit, path string) (*positionAdjuster, error) {
	if sourceCommit == targetCommit {
		// Trivial case, no changes to any file
		return &positionAdjuster{}, nil
	}

	output, err := readDiff(ctx, repo, sourceCommit, targetCommit, path)
	if err != nil {
		return nil, err
	}

	return newPositionAdjusterFromDiffOutput(output)
}

// readDiff returns the output git diff between the source and target commits for the given path.
func readDiff(ctx context.Context, repo *types.Repo, sourceCommit, targetCommit, path string) ([]byte, error) {
	cachedRepo, err := backend.CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	reader, err := git.ExecReader(ctx, *cachedRepo, []string{"diff", sourceCommit, targetCommit, "--", path})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}

// newPositionAdjusterFromReader creates a positionAdjuster directly from the output of a git diff
// command. The diff's original file is the source commit, and the new file is the target commit.
func newPositionAdjusterFromDiffOutput(output []byte) (*positionAdjuster, error) {
	if len(output) == 0 {
		// Trivial case, no changes to file
		return &positionAdjuster{}, nil
	}

	diff, err := diff.NewFileDiffReader(bytes.NewReader(output)).Read()
	if err != nil {
		return nil, err
	}

	return &positionAdjuster{hunks: diff.Hunks}, nil
}

func (pa *positionAdjuster) adjustRange(r lsp.Range) (lsp.Range, bool) {
	start, ok := pa.adjustPosition(r.Start)
	if !ok {
		return lsp.Range{}, false
	}

	end, ok := pa.adjustPosition(r.End)
	if !ok {
		return lsp.Range{}, false
	}

	return lsp.Range{Start: start, End: end}, true
}

// adjustPosition transforms the given position in the source commit to a position in the target
// commit. This method returns second boolean value indicating that the adjustment succeeded. If
// that particular line does not exist or has been edited in between the source and target commit,
// then adjustment will fail. The given position is assumed to be zero-indexed.
func (pa *positionAdjuster) adjustPosition(pos lsp.Position) (lsp.Position, bool) {
	// Find the index of the first hunk that starts after the target line and use the
	// previous hunk (if it exists) as the point of reference in `adjustPositionFromHunk`.
	// Note: LSP Positions are zero-indexed; the output of git diff is one-indexed.

	i := 0
	for i < len(pa.hunks) && int(pa.hunks[i].OrigStartLine) <= pos.Line+1 {
		i++
	}

	if i == 0 {
		// Trivial case, no changes before this line
		return pos, true
	}

	return adjustPositionFromHunk(pa.hunks[i-1], pos)
}

// adjustPositionFromHunk transforms the given position in the original file into a position
// in the new file according to the given git diff hunk. This parameter is expected to be the
// last such hunk in the diff between the original and the new file that does not begin after
// the given position in the original file.
func adjustPositionFromHunk(hunk *diff.Hunk, pos lsp.Position) (lsp.Position, bool) {
	// LSP Positions are zero-indexed; the output of git diff is one-indexed
	line := pos.Line + 1

	if line >= int(hunk.OrigStartLine+hunk.OrigLines) {
		// Hunk ends before this line, so we can simply adjust the line offset by the
		// relative difference between the line offsets in each file after this hunk.
		origEnd := int(hunk.OrigStartLine + hunk.OrigLines)
		newEnd := int(hunk.NewStartLine + hunk.NewLines)
		relativeDifference := newEnd - origEnd

		return lsp.Position{
			Line:      line + relativeDifference - 1,
			Character: pos.Character,
		}, true
	}

	// Create two fingers pointing at the first line of this hunk in each file. Then,
	// bump each of these cursors for every line in hunk body that is attributed
	// to the corresponding file.

	origFileOffset := int(hunk.OrigStartLine)
	newFileOffset := int(hunk.NewStartLine)

	for _, bodyLine := range strings.Split(string(hunk.Body), "\n") {
		// Bump original file offset unless it's an addition in the new file
		added := strings.HasPrefix(bodyLine, "+")
		if !added {
			origFileOffset++
		}

		// Bump new file offset unless it's a deletion of a line from the new file
		removed := strings.HasPrefix(bodyLine, "-")
		if !removed {
			newFileOffset++
		}

		// Keep skipping ahead in the original file until we hit our target line
		if origFileOffset-1 < line {
			continue
		}

		// This line exists in both files
		if !added && !removed {
			return lsp.Position{
				Line:      newFileOffset - 2,
				Character: pos.Character,
			}, true
		}

		// Fail the position adjustment. This particular line was either
		//   (1) edited;
		//   (2) removed in which case we can't point to it; or
		//   (3) added, in which case it hasn't been indexed.
		//
		// In all cases we don't want to return any results here as we
		// don't have enough information to give a precise result matching
		// the current query text.

		return lsp.Position{}, false
	}

	// This should never happen unless the git diff content is malformed. We know
	// the target line occurs within the hunk, but iteration of the hunk's body did
	// not contain enough lines attributed to the original file.
	panic("Malformed hunk body")
}
