package context

import (
	"bytes"
	"context"
	"strings"

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

type RepoContextFilter interface {
	Filter(chunks []FileChunkContext) []FileChunkContext
}

// NewCodyIgnoreFilter creates a new Filter that applies .cody/ignore rules from
// the given repositories to filter file paths. It reads the .cody/ignore file
// from each repo and parses it into an ignore.Matcher func that is stored in
// the returned Filter.
func NewCodyIgnoreFilter(ctx context.Context, client gitserver.Client, repos []types.RepoIDName) (RepoContextFilter, error) {
	f := &repoFilter{
		filters: make(map[types.RepoIDName]filterFunc),
	}
	for _, repo := range repos {
		ignoreFile, err := client.ReadFile(ctx, repo.Name, api.CommitID("HEAD"), codyIgnoreFile)
		if err != nil {
			// We do not ignore anything if the ignore file does not exist.
			if strings.Contains(err.Error(), "file does not exist") {
				continue
			}
			return nil, err
		}
		ig, err := ignore.ParseIgnoreFile(bytes.NewReader(ignoreFile))
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
