package search

import (
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// LazyCommit wraps a RawCommit and a DiffFetcher so that we can have a unified interface
// that makes all the information we need available without paying the cost of fetching
// diffs or parsing times when they're not unneeded.
type LazyCommit struct {
	*RawCommit
	diffFetcher *DiffFetcher
	diff        FormattedDiff
}

func (l *LazyCommit) AuthorDate() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RawCommit.AuthorDate))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

func (l *LazyCommit) CommitterDate() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RawCommit.CommitterDate))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

// RawDiff returns the diff exactly as returned by git diff-tree
func (l *LazyCommit) RawDiff() ([]byte, error) {
	return l.diffFetcher.FetchDiff(l.Hash)
}

// Diff fetches the diff, then formats it in the format used throughout our app
func (l *LazyCommit) Diff() (FormattedDiff, error) {
	if l.diff != "" {
		return l.diff, nil
	}

	rawDiff, err := l.RawDiff()
	if err != nil {
		return "", err
	}

	formattedDiff := FormatDiff(rawDiff)
	l.diff = formattedDiff
	return formattedDiff, nil
}

func (l *LazyCommit) ParentIDs() []api.CommitID {
	strs := strings.Split(string(l.ParentHashes), " ")
	commitIDs := make([]api.CommitID, 0, len(strs))
	for _, str := range strs {
		commitIDs = append(commitIDs, api.CommitID(str))
	}
	return commitIDs
}

func (l *LazyCommit) RefNames() []string {
	return strings.Split(string(l.RawCommit.RefNames), ", ")
}

func (l *LazyCommit) SourceRefs() []string {
	return strings.Split(string(l.RawCommit.SourceRefs), ", ")
}
