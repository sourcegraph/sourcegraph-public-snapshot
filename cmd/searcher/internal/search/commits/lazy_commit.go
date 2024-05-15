package commits

import (
	"context"
	"io"
	"time"

	godiff "github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// LazyCommit wraps a RawCommit and a DiffFetcher so that we can have a unified interface
// that makes all the information we need available without paying the cost of fetching
// diffs or parsing times when they're not unneeded.
type LazyCommit struct {
	*gitserver.WrappedCommit

	ctx context.Context

	// diff is the parsed output from the diff fetcher, cached here for performance
	diff        []*godiff.FileDiff
	diffFetcher *DiffFetcher

	// LowerBuf is a re-usable buffer for doing case-transformations on the fields of LazyCommit
	LowerBuf []byte
}

func (l *LazyCommit) AuthorDate() (time.Time, error) {
	return l.WrappedCommit.Author.Date, nil
}

func (l *LazyCommit) CommitterDate() (time.Time, error) {
	return l.WrappedCommit.Committer.Date, nil
}

// RawDiff returns the diff exactly as returned by git diff-tree
func (l *LazyCommit) RawDiff() (*gitserver.DiffFileIterator, error) {
	return l.diffFetcher.Fetch([]byte(l.ID))
}

// Diff fetches the diff, then parses it with go-diff, caching the result
func (l *LazyCommit) Diff() ([]*godiff.FileDiff, error) {
	if l.diff != nil {
		return l.diff, nil
	}

	rawDiff, err := l.RawDiff()
	if err != nil {
		return nil, err
	}

	defer rawDiff.Close()

	diff := []*godiff.FileDiff{}
	for {
		d, err := rawDiff.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		diff = append(diff, d)
	}

	l.diff = diff
	return diff, nil
}

func (l *LazyCommit) ParentIDs() ([]api.CommitID, error) {
	return l.Parents, nil
}

// func (l *LazyCommit) RefNames() []string {
// 	return strings.Split(utf8String(l.RawCommit.RefNames), ", ")
// }

// func (l *LazyCommit) SourceRefs() []string {
// 	return strings.Split(utf8String(l.RawCommit.SourceRefs), ", ")
// }

func (l *LazyCommit) ModifiedFiles() []string {
	// TODO: This is different to before, we used to also include the file name
	// at the PREVIOUS commit for files that were renamed.
	return l.WrappedCommit.ChangedFiles
	// files := make([]string, 0, len(l.WrappedCommit.ChangedFiles)/2)
	// i := 0
	// for i < len(l.RawCommit.ModifiedFiles) {
	// 	if len(l.RawCommit.ModifiedFiles[i]) == 0 {
	// 		// SAFETY: don't trust input
	// 		return files
	// 	}
	// 	switch l.RawCommit.ModifiedFiles[i][0] {
	// 	case 'R':
	// 		// SAFETY: don't assume that we have the right number of things
	// 		if i+2 >= len(l.RawCommit.ModifiedFiles) {
	// 			return files
	// 		}
	// 		// A rename entry will be followed by two file names
	// 		files = append(files, utf8String(l.RawCommit.ModifiedFiles[i+1]))
	// 		files = append(files, utf8String(l.RawCommit.ModifiedFiles[i+2]))
	// 		i += 3
	// 	default:
	// 		// SAFETY: don't assume that we have the right number of things
	// 		if i+1 >= len(l.RawCommit.ModifiedFiles) {
	// 			return files
	// 		}
	// 		// Any entry that is not a rename entry will only be followed by one file name
	// 		files = append(files, utf8String(l.RawCommit.ModifiedFiles[i+1]))
	// 		i += 2
	// 	}
	// }
	// return files
}
