package context

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/sourcegraph/zoekt/ignore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const codyIgnoreFile = ".cody/ignore"

type filterFunc func(string) bool
type repoFilter struct {
	filters map[types.RepoIDName]filterFunc
}

type RepoContentFilter interface {
	Filter(chunks []FileChunkContext) []FileChunkContext
}

// NewCodyIgnoreFilter creates a new RepoContentFilter that filters out
// content based on the .cody/ignore file at the head of the default branch
// for the given repositories. If no .cody/ignore file exists, no filtering is done.
func NewCodyIgnoreFilter(ctx context.Context, client gitserver.Client, repos []types.RepoIDName) (RepoContentFilter, error) {
	f := &repoFilter{
		filters: make(map[types.RepoIDName]filterFunc),
	}
	for _, repo := range repos {
		_, commit, err := client.GetDefaultBranch(ctx, repo.Name, true)
		if err != nil {
			return nil, err
		}
		// this is an empty repo, there won't be anything to filter
		if commit == "" {
			continue
		}
		ignoreFileBytes, err := getIgnoreFileBytes(ctx, client, repo, commit)
		if err != nil {
			// We do not ignore anything if the ignore file does not exist.
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		ig, err := ignore.ParseIgnoreFile(bytes.NewReader(ignoreFileBytes))
		if err != nil {
			return nil, err
		}
		f.filters[repo] = ig.Match
	}

	return f, nil
}

// Filter applies the ignore rules to the given file chunks,
// returning only those that do not match any ignore rules.
func (f *repoFilter) Filter(chunks []FileChunkContext) []FileChunkContext {
	filtered := make([]FileChunkContext, 0, len(chunks))
	for _, chunk := range chunks {
		ignore, ok := f.filters[types.RepoIDName{ID: chunk.RepoID, Name: chunk.RepoName}]
		if !ok || !ignore(chunk.Path) {
			filtered = append(filtered, chunk)
			continue
		}
	}
	return filtered
}

func getIgnoreFileBytes(ctx context.Context, client gitserver.Client, repo types.RepoIDName, commit api.CommitID) ([]byte, error) {
	fr, err := client.NewFileReader(
		ctx,
		repo.Name,
		commit,
		codyIgnoreFile,
	)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	return io.ReadAll(fr)
}
