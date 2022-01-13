package search

import (
	"archive/tar"
	"bytes"
	"context"
	"strings"

	"github.com/google/zoekt/ignore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// This matches the default file size limit in Zoekt
// https://github.com/sourcegraph/zoekt/blob/a2c843bdb1bffcfaf674034ddfd35403f90a70ac/build/builder.go#L240
const maxFileSize = 2 << 20

// NewFilter is a wrapper around newIgnoreMatcher.
func NewFilter(ctx context.Context, repo api.RepoName, commit api.CommitID) (store.FilterFunc, error) {
	ig, err := newIgnoreMatcher(ctx, repo, commit)
	if err != nil {
		return nil, err
	}
	return func(header *tar.Header) bool {
		if header.Size > maxFileSize {
			return true
		}
		return ig.Match(header.Name)
	}, nil
}

// newIgnoreMatcher calls gitserver to retrieve the ignore-file.
// If the file doesn't exist we return an empty ignore.Matcher.
func newIgnoreMatcher(ctx context.Context, repo api.RepoName, commit api.CommitID) (*ignore.Matcher, error) {
	ignoreFile, err := git.ReadFile(ctx, repo, commit, ignore.IgnoreFile, 0, nil)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") {
			return &ignore.Matcher{}, nil
		}
		return nil, err
	}
	return ignore.ParseIgnoreFile(bytes.NewReader(ignoreFile))
}
