package resolvers

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// adjustPosition transforms the given position in the source commit to a position in the target
// commit. This function returns a nil position if that particular line does not exist or has been
// edited in between the source and target commit.
func adjustPosition(
	ctx context.Context,
	repo *types.Repo,
	sourceCommit string,
	targetCommit string,
	path string,
	pos lsp.Position,
) (*lsp.Position, error) {
	if sourceCommit == targetCommit {
		// Trivial case, we have this exact commit indexed
		return &pos, nil
	}

	cachedRepo, err := backend.CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	reader, err := git.ExecReader(
		ctx,
		*cachedRepo,
		[]string{"diff", sourceCommit, targetCommit, "--", path},
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Non-trivial case, adjust the line offset based on diff hunks.
	return adjustPositionFromDiff(reader, pos)
}

func adjustRange(
	ctx context.Context,
	repo *types.Repo,
	sourceCommit string,
	targetCommit string,
	path string,
	lspRange lsp.Range,
) (*lsp.Range, error) {
	start, err := adjustPosition(ctx, repo, sourceCommit, targetCommit, path, lspRange.Start)
	if err != nil || start == nil {
		return nil, err
	}

	end, err := adjustPosition(ctx, repo, sourceCommit, targetCommit, path, lspRange.End)
	if err != nil || end == nil {
		return nil, err
	}

	return &lsp.Range{
		Start: lsp.Position{Line: int(start.Line), Character: int(start.Character)},
		End:   lsp.Position{Line: int(end.Line), Character: int(end.Character)},
	}, nil
}

// adjustPositionFromDiff transforms the given position in the *original file* into a position
// in the *new file* according to the git diff output contained in the given reader. This
// function returns a nil position if the line has been edited or does not exist in the new
// file. The given position is assumed to be zero-indexed.
func adjustPositionFromDiff(reader io.Reader, pos lsp.Position) (*lsp.Position, error) {
	// LSP Positions are zero-indexed; the output of git diff is one-indexed
	line := pos.Line + 1

	// Extract hunks from the git diff output
	diff, err := diff.NewFileDiffReader(reader).Read()
	if err != nil {
		return nil, err
	}
	hunks := diff.Hunks

	// Find the index of the first hunk that starts after the target line and use the
	// previous hunk (if it exists) as the point of reference in `adjustPositionFromHunk`.

	i := 0
	for i < len(hunks) && int(hunks[i].OrigStartLine) <= line {
		i++
	}

	if i == 0 {
		return adjustPositionFromHunk(nil, pos)
	}
	return adjustPositionFromHunk(hunks[i-1], pos)
}

// adjustPositionFromHunk transforms the given position in the *original file* into a position
// in the *new file* according to the given git diff hunk. This parameter is expected to be the
// *last* such hunk in the diff between the original and the new file that does not begin after
// the given position in the original file. The given position is assumed to be zero-indexed.
func adjustPositionFromHunk(hunk *diff.Hunk, pos lsp.Position) (*lsp.Position, error) {
	if hunk == nil {
		// No hunk before this line, so no line offset
		return &pos, nil
	}

	// LSP Positions are zero-indexed; the output of git diff is one-indexed
	line := pos.Line + 1

	if line >= int(hunk.OrigStartLine+hunk.OrigLines) {
		// Hunk ends before this line, so we can simply adjust the line offset by the
		// relative difference between the line offsets in each file after this hunk.
		relativeDifference := int(hunk.NewStartLine+hunk.NewLines) - int(hunk.OrigStartLine+hunk.OrigLines)
		return &lsp.Position{Line: line + relativeDifference - 1, Character: pos.Character}, nil
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
		if int(origFileOffset-1) < line {
			continue
		}

		// This line exists in both files
		if !added && !removed {
			return &lsp.Position{Line: newFileOffset - 2, Character: pos.Character}, nil
		}

		// Fail the position adjustment. This particular line
		//   (1) edited;
		//   (2) removed in which case we can't point to it; or
		//   (3) added, in which case it hasn't been indexed.
		//
		// In all cases we don't want to return any results here as we
		// don't have enough information to give a precise result matching
		// the current query text.

		return nil, nil
	}

	// This should never happen unless the git diff content is malformed.
	// We know the target line occurs within the hunk, but iteration of the
	// hunk's body did not contain enough lines attributed to the original
	// file.
	return nil, fmt.Errorf("malformed git diff hunk")
}
