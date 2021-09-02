package highlight

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func fetchContent(ctx context.Context, fm *result.FileMatch) (content []byte, err error) {
	var contentOnce sync.Once
	contentOnce.Do(func() {
		content, err = git.ReadFile(ctx, fm.Repo.Name, fm.CommitID, fm.Path, 0)
	})
	if err != nil {
		return nil, err
	}
	return content, nil
}

// Decorate file content given FileMatch.
func Decorate(ctx context.Context, fm *result.FileMatch) (string, error) {
	content, err := fetchContent(ctx, fm)
	if err != nil {
		return "", err
	}
	result, aborted, err := Code(ctx, Params{
		Content:            content,
		Filepath:           fm.Path,
		DisableTimeout:     false, // TODO: when false, sets 3 second timeout
		IsLightTheme:       false, // unused
		HighlightLongLines: false, // TODO: limits to 2000 by default
		SimulateTimeout:    false, // for test
		Metadata: Metadata{ // for logging
			RepoName: string(fm.Repo.Name),
			Revision: string(fm.CommitID),
		},
	})
	if err != nil {
		return "", err
	}
	if aborted {
		// TODO: make Decorate return a value that indicates whether it plaintext HTML,
		// or decide whether to return nothing.

		// code decoration aborted, returns plaintext HTML.
		return string(result), nil
	}
	return string(result), nil
}
