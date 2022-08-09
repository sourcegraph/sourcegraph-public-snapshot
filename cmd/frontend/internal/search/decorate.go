package search

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	stream "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// decorateFileChunksHTML returns decorated file hunks given a file match.
func decorateFileChunksHTML(ctx context.Context, db database.DB, contextLines int, cm *stream.EventContentMatch) error {
	response, err := decorateFileHTML(ctx, db, api.RepoName(cm.Repository), api.CommitID(cm.Commit), cm.Path)
	if err != nil {
		return errors.Wrap(err, "highlight file")
	}

	lines, err := response.SplitHighlightedLines(true)
	if err != nil {
		return errors.Wrap(err, "split highlighted lines")
	}

	// a closure over lines that allows to splice line ranges.
	spliceRows := func(lineStart, lineEnd int) []string {
		if lineStart < 0 {
			lineStart = 0
		}
		if lineEnd > len(lines) {
			lineEnd = len(lines)
		}
		if lineStart > lineEnd {
			lineStart = 0
			lineEnd = 0
		}
		tableRows := make([]string, 0, lineEnd-lineStart)
		for _, line := range lines[lineStart:lineEnd] {
			tableRows = append(tableRows, string(line))
		}
		return tableRows
	}

	for i := range cm.ChunkMatches {
		chunk := cm.ChunkMatches[i]
		rows := spliceRows(chunk.ContentStart.Line-contextLines, chunk.ContentStart.Line+strings.Count(chunk.Content, "\n")+1+contextLines)
		chunk.HTMLDecoratedContent = strings.Join(rows, "\n")
		cm.ChunkMatches[i] = chunk
	}
	return nil
}

// decorateFileHTML returns decorated HTML rendering of file content. If
// successful and within bounds of timeout and line size, it returns HTML marked
// up with highlight classes. In other cases, it returns plaintext HTML.
func decorateFileHTML(ctx context.Context, db database.DB, repo api.RepoName, commit api.CommitID, path string) (*highlight.HighlightedCode, error) {
	content, err := fetchContent(ctx, db, repo, commit, path)
	if err != nil {
		return nil, err
	}

	highlightResponse, _, err := highlight.Code(ctx, highlight.Params{
		Content:            content,
		Filepath:           path,
		DisableTimeout:     false, // use default 3 second timeout
		HighlightLongLines: false, // use default 2000 character line count
		Metadata: highlight.Metadata{ // for logging
			RepoName: string(repo),
			Revision: string(commit),
		},
	})
	return highlightResponse, err
}

func fetchContent(ctx context.Context, db database.DB, repo api.RepoName, commit api.CommitID, path string) (content []byte, err error) {
	content, err = gitserver.NewClient(db).ReadFile(ctx, repo, commit, path, authz.DefaultSubRepoPermsChecker)
	if err != nil {
		return nil, err
	}
	return content, nil
}
