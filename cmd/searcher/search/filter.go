package search

import (
	"archive/tar"
	"bytes"
	"context"
	"strings"

	"github.com/google/zoekt/ignore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// This matches the default file size limit in Zoekt
// https://github.com/sourcegraph/zoekt/blob/a2c843bdb1bffcfaf674034ddfd35403f90a70ac/build/builder.go#L240
const maxFileSize = 2 << 20

// NewFilter calls gitserver to retrieve the ignore-file. If the file doesn't
// exist we return an empty ignore.Matcher.
func NewFilter(ctx context.Context, db database.DB, repo api.RepoName, commit api.CommitID) (store.FilterFunc, error) {
	ignoreFile, err := git.ReadFile(ctx, db, repo, commit, ignore.IgnoreFile, 0, nil)
	if err != nil {
		// We do not ignore anything if the ignore file does not exist.
		if strings.Contains(err.Error(), "file does not exist") {
			return func(*tar.Header) bool {
				return false
			}, nil
		}
		return nil, err
	}

	ig, err := ignore.ParseIgnoreFile(bytes.NewReader(ignoreFile))
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
