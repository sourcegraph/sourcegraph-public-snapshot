package search

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// MatchedCommit are the portions of a commit that match a query
type MatchedCommit struct {
	// Message is the set of ranges of the commit message that were matched
	Message result.Ranges

	// Diff is the set of files deltas that have matches in the parsed diff.
	// The key of the map is the index of the delta in the diff.
	Diff map[int]MatchedFileDiff
}

// Merge merges another CommitHighlights into this one, returning the result.
func (c MatchedCommit) Merge(other MatchedCommit) MatchedCommit {
	c.Message = c.Message.Merge(other.Message)

	if c.Diff == nil {
		c.Diff = other.Diff
	} else {
		for i, fdh := range other.Diff {
			c.Diff[i] = c.Diff[i].Merge(fdh)
		}
	}

	return c
}

// ConstrainToMatched constrains a MatchedCommit by deleting any match ranges for file
// diffs not included in the provided matchedFileDiffs.
func (c *MatchedCommit) ConstrainToMatched(matchedFileDiffs map[int]struct{}) {
	for i := range c.Diff {
		if _, ok := matchedFileDiffs[i]; !ok {
			delete(c.Diff, i)
		}
	}
}

type MatchedFileDiff struct {
	OldFile      result.Ranges
	NewFile      result.Ranges
	MatchedHunks map[int]MatchedHunk
}

func (f MatchedFileDiff) Merge(other MatchedFileDiff) MatchedFileDiff {
	f.OldFile = append(f.OldFile, other.OldFile...)
	sort.Sort(f.OldFile)

	f.NewFile = append(f.NewFile, other.NewFile...)
	sort.Sort(f.NewFile)

	if f.MatchedHunks == nil {
		f.MatchedHunks = other.MatchedHunks
	} else {
		for i, hh := range other.MatchedHunks {
			f.MatchedHunks[i] = f.MatchedHunks[i].Merge(hh)
		}
	}
	return f
}

type MatchedHunk struct {
	MatchedLines map[int]result.Ranges
}

func (h MatchedHunk) Merge(other MatchedHunk) MatchedHunk {
	if h.MatchedLines == nil {
		h.MatchedLines = other.MatchedLines
	} else {
		for i, lh := range other.MatchedLines {
			h.MatchedLines[i] = h.MatchedLines[i].Merge(lh)
		}
	}
	return h
}
