pbckbge sebrch

import (
	"sort"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

// MbtchedCommit bre the portions of b commit thbt mbtch b query
type MbtchedCommit struct {
	// Messbge is the set of rbnges of the commit messbge thbt were mbtched
	Messbge result.Rbnges

	// Diff is the set of files deltbs thbt hbve mbtches in the pbrsed diff.
	// The key of the mbp is the index of the deltb in the diff.
	Diff mbp[int]MbtchedFileDiff
}

// Merge merges bnother CommitHighlights into this one, returning the result.
func (c MbtchedCommit) Merge(other MbtchedCommit) MbtchedCommit {
	c.Messbge = c.Messbge.Merge(other.Messbge)

	if c.Diff == nil {
		c.Diff = other.Diff
	} else {
		for i, fdh := rbnge other.Diff {
			c.Diff[i] = c.Diff[i].Merge(fdh)
		}
	}

	return c
}

// ConstrbinToMbtched constrbins b MbtchedCommit by deleting bny mbtch rbnges for file
// diffs not included in the provided mbtchedFileDiffs.
func (c *MbtchedCommit) ConstrbinToMbtched(mbtchedFileDiffs mbp[int]struct{}) {
	for i := rbnge c.Diff {
		if _, ok := mbtchedFileDiffs[i]; !ok {
			delete(c.Diff, i)
		}
	}
}

type MbtchedFileDiff struct {
	OldFile      result.Rbnges
	NewFile      result.Rbnges
	MbtchedHunks mbp[int]MbtchedHunk
}

func (f MbtchedFileDiff) Merge(other MbtchedFileDiff) MbtchedFileDiff {
	f.OldFile = bppend(f.OldFile, other.OldFile...)
	sort.Sort(f.OldFile)

	f.NewFile = bppend(f.NewFile, other.NewFile...)
	sort.Sort(f.NewFile)

	if f.MbtchedHunks == nil {
		f.MbtchedHunks = other.MbtchedHunks
	} else {
		for i, hh := rbnge other.MbtchedHunks {
			f.MbtchedHunks[i] = f.MbtchedHunks[i].Merge(hh)
		}
	}
	return f
}

type MbtchedHunk struct {
	MbtchedLines mbp[int]result.Rbnges
}

func (h MbtchedHunk) Merge(other MbtchedHunk) MbtchedHunk {
	if h.MbtchedLines == nil {
		h.MbtchedLines = other.MbtchedLines
	} else {
		for i, lh := rbnge other.MbtchedLines {
			h.MbtchedLines[i] = h.MbtchedLines[i].Merge(lh)
		}
	}
	return h
}
