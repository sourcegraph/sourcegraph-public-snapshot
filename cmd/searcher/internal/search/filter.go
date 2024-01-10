package search

import (
	"archive/tar"
	"bytes"
	"context"
	"hash"
	"io"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/sourcegraph/zoekt/ignore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewFilter calls gitserver to retrieve the ignore-file. If the file doesn't
// exist we return an empty ignore.Matcher.
func NewFilter(ctx context.Context, client gitserver.Client, repo api.RepoName, commit api.CommitID) (FilterFunc, error) {
	ignoreFile, err := client.ReadFile(ctx, repo, commit, ignore.IgnoreFile)
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

func newSearchableFilter(c *schema.SiteConfiguration) *searchableFilter {
	return &searchableFilter{
		SearchLargeFiles: c.SearchLargeFiles,
	}
}

// searchableFilter contains logic for what should and should not be stored in
// the store.
type searchableFilter struct {
	// CommitIgnore filters out files that should not appear at all based on
	// the commit. This does not contribute to HashKey and is only set once we
	// start fetching the archive. This is since it is part of the state of
	// the commit, and not derivable from the request.
	//
	// See NewFilter function above.
	CommitIgnore FilterFunc

	// SearchLargeFiles is a list of globs for files were we do not respect
	// fileSizeMax. It comes from the site configuration search.largeFiles.
	SearchLargeFiles []string
}

// Ignore returns true if the file should not appear at all when searched. IE
// is excluded for both filename and content searches.
//
// Note: This function relies on CommitIgnore being set by NewFilter. Not
// calling NewFilter indicates a bug and as such will panic.
func (f *searchableFilter) Ignore(hdr *tar.Header) bool {
	return f.CommitIgnore(hdr)
}

// SkipContent returns true if we should not include the content of the file
// in the search. This means you can still find the file by filename, but not
// by its contents.
func (f *searchableFilter) SkipContent(hdr *tar.Header) bool {
	// We do not search the content of large files unless they are
	// allowed.
	if hdr.Size <= maxFileSize {
		return false
	}

	// A pattern match will override preceding pattern matches.
	for i := len(f.SearchLargeFiles) - 1; i >= 0; i-- {
		pattern := strings.TrimSpace(f.SearchLargeFiles[i])
		negated, validatedPattern := checkIsNegatePattern(pattern)
		if m, _ := doublestar.PathMatch(validatedPattern, hdr.Name); m {
			if negated {
				return true // overrides any preceding inclusion patterns
			} else {
				return false // overrides any preceding exclusion patterns
			}
		}
	}

	return true
}

func checkIsNegatePattern(pattern string) (bool, string) {
	negate := "!"

	// if negated then strip prefix meta character which identifies negated filter pattern
	if strings.HasPrefix(pattern, negate) {
		return true, pattern[len(negate):]
	}

	return false, pattern
}

// HashKey will write the input of the filter to h.
//
// This is used as part of the key of what is stored on disk, such that if the
// configuration changes we invalidated the cache.
func (f *searchableFilter) HashKey(h hash.Hash) {
	_, _ = io.WriteString(h, "\x00SearchLargeFiles")
	for _, p := range f.SearchLargeFiles {
		_, _ = h.Write([]byte{0})
		_, _ = io.WriteString(h, p)
	}
}
