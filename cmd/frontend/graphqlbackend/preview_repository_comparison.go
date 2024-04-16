package graphqlbackend

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PreviewRepositoryComparisonResolver interface {
	RepositoryComparisonInterface
}

// NewPreviewRepositoryComparisonResolver is a convenience function to get a preview diff from a repo, given a base rev and the git patch.
func NewPreviewRepositoryComparisonResolver(ctx context.Context, db database.DB, client gitserver.Client, repo *RepositoryResolver, baseRev string, patch []byte) (*previewRepositoryComparisonResolver, error) {
	args := &RepositoryCommitArgs{Rev: baseRev}
	commit, err := repo.Commit(ctx, args)
	if err != nil {
		return nil, err
	}
	if commit == nil {
		return nil, &gitdomain.RevisionNotFoundError{
			Repo: api.RepoName(repo.Name()),
			Spec: baseRev,
		}
	}
	return &previewRepositoryComparisonResolver{
		db:     db,
		client: client,
		repo:   repo,
		commit: commit,
		patch:  patch,
	}, nil
}

type previewRepositoryComparisonResolver struct {
	db     database.DB
	client gitserver.Client
	repo   *RepositoryResolver
	commit *GitCommitResolver
	patch  []byte
}

// Type guard.
var _ RepositoryComparisonInterface = &previewRepositoryComparisonResolver{}

func (r *previewRepositoryComparisonResolver) ToPreviewRepositoryComparison() (PreviewRepositoryComparisonResolver, bool) {
	return r, true
}

func (r *previewRepositoryComparisonResolver) ToRepositoryComparison() (*RepositoryComparisonResolver, bool) {
	return nil, false
}

func (r *previewRepositoryComparisonResolver) BaseRepository() *RepositoryResolver {
	return r.repo
}

func (r *previewRepositoryComparisonResolver) FileDiffs(_ context.Context, args *FileDiffsConnectionArgs) (FileDiffConnection, error) {
	return NewFileDiffConnectionResolver(r.db, r.client, r.commit, r.commit, args, fileDiffConnectionCompute(r.patch), previewNewFile), nil
}

func fileDiffConnectionCompute(patch []byte) func(ctx context.Context, args *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
	var (
		once        sync.Once
		fileDiffs   []*diff.FileDiff
		afterIdx    int32
		hasNextPage bool
		err         error
	)
	return func(ctx context.Context, args *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
		once.Do(func() {
			if args.After != nil {
				parsedIdx, err := strconv.ParseInt(*args.After, 0, 32)
				if err != nil {
					return
				}
				if parsedIdx < 0 {
					parsedIdx = 0
				}
				afterIdx = int32(parsedIdx)
			}
			totalAmount := afterIdx
			if args.First != nil {
				totalAmount += *args.First
			}

			dr := diff.NewMultiFileDiffReader(bytes.NewReader(patch))
			for {
				var fileDiff *diff.FileDiff
				fileDiff, err = dr.ReadFile()
				if err == io.EOF {
					err = nil
					break
				}
				if err != nil {
					return
				}
				fileDiffs = append(fileDiffs, fileDiff)
				if len(fileDiffs) == int(totalAmount) {
					// Check for hasNextPage.
					_, err = dr.ReadFile()
					if err != nil && err != io.EOF {
						return
					}
					if err == io.EOF {
						err = nil
					} else {
						hasNextPage = true
					}
					break
				}
			}
		})
		return fileDiffs, afterIdx, hasNextPage, err
	}
}

func previewNewFile(db database.DB, r *fileDiffResolver) FileResolver {
	fileStat := CreateFileInfo(r.FileDiff.NewName, false)
	return NewVirtualFileResolver(fileStat, fileDiffVirtualFileContent(r), VirtualFileResolverOptions{
		// TODO: Add view in webapp to render full preview files.
		URL: "",
	})
}

func fileDiffVirtualFileContent(r *fileDiffResolver) FileContentFunc {
	var (
		once       sync.Once
		newContent string
		err        error
	)
	return func(ctx context.Context) (string, error) {
		once.Do(func() {
			var oldContent string
			if oldFile := r.OldFile(); oldFile != nil {
				var err error
				oldContent, err = r.OldFile().Content(ctx, &GitTreeContentPageArgs{})
				if err != nil {
					return
				}
			}
			newContent, err = applyPatch(oldContent, r.FileDiff)
		})
		return newContent, err
	}
}

// applyPatch takes the contents of a file and a file diff to apply to it. It
// returns the patched file content.
func applyPatch(fileContent string, fileDiff *diff.FileDiff) (string, error) {
	if diffPathOrNull(fileDiff.NewName) == nil {
		// the file was deleted, no need to do costly computation.
		return "", nil
	}

	// Capture if the original file content had a final newline.
	origHasFinalNewline := strings.HasSuffix(fileContent, "\n")

	// All the lines of the original file.
	var contentLines []string
	// For empty files, don't split otherwise we end up with an empty ghost line.
	if len(fileContent) > 0 {
		// Trim a potential final newline so that we don't end up with an empty
		// ghost line at the end.
		contentLines = strings.Split(strings.TrimSuffix(fileContent, "\n"), "\n")
	}
	// The new lines of the file.
	newContentLines := make([]string, 0)
	// Track whether the last hunk seen ended with an addition and the hunk indicated
	// the new file potentially has a final new line, if the last hunk was at the
	// end of the file. This is rechecked further down.
	lastHunkHadNewlineInLastAddition := false
	// Track the current line index to be processed. Line is 1-indexed.
	var currentLine int32 = 1
	isLastHunk := func(i int) bool {
		return i == len(fileDiff.Hunks)-1
	}
	// Assumes the hunks are sorted by ascending lines.
	for i, hunk := range fileDiff.Hunks {
		// Detect holes. If we are not at the start, or the hunks are not fully consecutive,
		// we need to fill up the lines in between.
		if hunk.OrigStartLine != 0 && hunk.OrigStartLine != currentLine {
			originalLines := contentLines[currentLine-1 : hunk.OrigStartLine-1]
			newContentLines = append(newContentLines, originalLines...)
			// If we add the first 10 lines, we are now at line 11.
			currentLine += int32(len(originalLines))
		}
		// Iterate over all the hunk lines. Trim a potential final newline, so that
		// we don't end up with a ghost line in the slice.
		hunkLines := strings.Split(strings.TrimSuffix(string(hunk.Body), "\n"), "\n")
		hunkHasFinalNewline := strings.HasSuffix(string(hunk.Body), "\n")
		if isLastHunk(i) && hunkHasFinalNewline {
			lastHunkHadNewlineInLastAddition = true
		}
		for _, line := range hunkLines {
			switch {
			case strings.HasPrefix(line, " "):
				newContentLines = append(newContentLines, contentLines[currentLine-1])
				currentLine++
			case strings.HasPrefix(line, "-"):
				currentLine++
			case strings.HasPrefix(line, "+"):
				// Append the line, stripping off the diff signifier at the beginning.
				newContentLines = append(newContentLines, line[1:])
			default:
				return "", errors.Newf("malformed patch, expected hunk lines to start with ' ', '-', or '+' but got %q", line)
			}
		}
	}

	// If we are not at the end of the file, append remaining lines from original content.
	// Example:
	// The file had 20 lines, origLines = 20.
	// We only had a hunk go until line 14 of the original content.
	// currentLine is 15 now.
	// So we need to add all the remaining lines, from 15-20.
	if origLines := int32(len(contentLines)); origLines > 0 && origLines != currentLine-1 {
		newContentLines = append(newContentLines, contentLines[currentLine-1:]...)
		content := strings.Join(newContentLines, "\n")
		if origHasFinalNewline {
			content += "\n"
		}
		return content, nil
	}

	content := strings.Join(newContentLines, "\n")
	// If we are here, that means that a hunk covered the end of the file.
	// If the very last hunk ends with a deletion we're done.
	// If we ended with a new line, we need to make sure that we correctly reflect
	// the newline state of that.
	// If a newline IS present in the new content, we need to apply a final newline,
	// otherwise that means the file has no newline at the end.
	if lastHunkHadNewlineInLastAddition {
		// If the file has a final newline character, we need to append it again.
		content += "\n"
	}

	return content, nil
}
