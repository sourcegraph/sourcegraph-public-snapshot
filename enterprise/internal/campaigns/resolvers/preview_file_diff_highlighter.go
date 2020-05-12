package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
)

type previewFileDiffHighlighter struct {
	previewFileDiffResolver *previewFileDiffResolver
	highlightedBase         []string
	highlightedHead         []string
	highlightOnce           sync.Once
	highlightErr            error
	highlightAborted        bool
}

func (r *previewFileDiffHighlighter) Highlight(ctx context.Context, args *graphqlbackend.HighlightArgs) ([]string, []string, bool, error) {
	r.highlightOnce.Do(func() {
		if oldFile := r.previewFileDiffResolver.OldFile(); oldFile != nil {
			binary, err := oldFile.Binary(ctx)
			if err != nil {
				r.highlightErr = err
				return
			}
			if !binary {
				highlightedBaseTable, err := oldFile.Highlight(ctx, &graphqlbackend.HighlightArgs{
					DisableTimeout:     args.DisableTimeout,
					HighlightLongLines: args.HighlightLongLines,
					IsLightTheme:       args.IsLightTheme,
				})
				if err != nil {
					r.highlightErr = err
					return
				}
				if highlightedBaseTable.Aborted() {
					r.highlightAborted = true
				}
				r.highlightedBase, r.highlightErr = highlight.ParseLinesFromHighlight(highlightedBaseTable.HTML())
				if r.highlightErr != nil {
					return
				}
			}
		}

		if newPath := r.previewFileDiffResolver.NewPath(); newPath != nil {
			var content string
			if oldFile := r.previewFileDiffResolver.OldFile(); oldFile != nil {
				var err error
				content, err = r.previewFileDiffResolver.OldFile().Content(ctx)
				if err != nil {
					r.highlightErr = err
					return
				}
				binary, err := oldFile.Binary(ctx)
				if err != nil {
					r.highlightErr = err
					return
				}
				if binary {
					return
				}
			}
			newContent := applyPatch(content, r.previewFileDiffResolver.fileDiff)
			if highlight.IsBinary([]byte(newContent)) {
				return
			}
			highlightedHeadTable, aborted, err := highlight.Code(ctx, highlight.Params{
				Content:  []byte(newContent),
				Filepath: *newPath,
				Metadata: highlight.Metadata{
					RepoName: r.previewFileDiffResolver.commit.Repository().Name(),
					Revision: string(r.previewFileDiffResolver.commit.OID()),
				},
				DisableTimeout:     args.DisableTimeout,
				HighlightLongLines: args.HighlightLongLines,
				IsLightTheme:       args.IsLightTheme,
			})
			if err != nil {
				r.highlightErr = err
				return
			}
			if aborted {
				r.highlightAborted = true
			}
			r.highlightedHead, r.highlightErr = highlight.ParseLinesFromHighlight(string(highlightedHeadTable))
		}
	})
	return r.highlightedBase, r.highlightedHead, r.highlightAborted, r.highlightErr
}

func applyPatch(fileContent string, fileDiff *diff.FileDiff) string {
	contentLines := strings.Split(fileContent, "\n")
	newContentLines := make([]string, 0)
	var lastLine int32 = 1
	// Assumes the hunks are sorted by ascending lines.
	for _, hunk := range fileDiff.Hunks {
		// Detect holes.
		if hunk.OrigStartLine != 0 && hunk.OrigStartLine != lastLine {
			originalLines := contentLines[lastLine-1 : hunk.OrigStartLine-1]
			newContentLines = append(newContentLines, originalLines...)
			lastLine += int32(len(originalLines))
		}
		hunkLines := strings.Split(string(hunk.Body), "\n")
		for _, line := range hunkLines {
			switch {
			case line == "":
				// Skip
			case strings.HasPrefix(line, "-"):
				lastLine++
			case strings.HasPrefix(line, "+"):
				newContentLines = append(newContentLines, line[1:])
			default:
				newContentLines = append(newContentLines, contentLines[lastLine-1])
				lastLine++
			}
		}
	}
	// Append remaining lines from original file.
	if origLines := int32(len(contentLines)); origLines > 0 && origLines != lastLine {
		newContentLines = append(newContentLines, contentLines[lastLine-1:]...)
	}
	return strings.Join(newContentLines, "\n")
}
