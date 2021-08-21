package search

import (
	"sort"
)

type CommitHighlights struct {
	Diff    FileHighlights
	Message Ranges
	// TODO we could fairly easily expand highlights to include fields like Author and Committer
	// Author    Ranges
	// Committer Ranges
}

func (c *CommitHighlights) Merge(other *CommitHighlights) {
	if other == nil {
		return
	}

	if c.Diff == nil {
		c.Diff = other.Diff
	} else {
		c.Diff = c.Diff.Merge(other.Diff)
	}

	c.Message = append(c.Message, other.Message...)
	c.Message.MergeOverlapping()
}

func (c *CommitHighlights) AddDiffLineMatches(fileNum, hunkNum, lineNum int, ranges Ranges) {
	if c.Diff == nil {
		c.Diff = make(FileHighlights)
	}

	if _, ok := c.Diff[fileNum]; !ok {
		c.Diff[fileNum] = FileHighlight{
			Hunk: make(HunkHighlights),
		}
	}

	if c.Diff[fileNum].Hunk[hunkNum] == nil {
		c.Diff[fileNum].Hunk[hunkNum] = make(LineHighlights)
	}

	c.Diff[fileNum].Hunk[hunkNum][lineNum] = append(c.Diff[fileNum].Hunk[hunkNum][lineNum], ranges...)
}

func (c *CommitHighlights) AddFileNameHighlights(fileNum int, oldFile, newFile Ranges) {
	fh, ok := c.Diff[fileNum]
	if !ok {
		fh = FileHighlight{}
		c.Diff[fileNum] = fh
	}
	fh.OldFile = append(fh.OldFile, oldFile...)
	fh.NewFile = append(fh.NewFile, newFile...)
}

type FileHighlights map[int]FileHighlight

func (f FileHighlights) Merge(other FileHighlights) FileHighlights {
	if len(other) == 0 {
		return f
	}

	if len(f) == 0 {
		return other
	}

	for i, newHunkHighlight := range other {
		oldHunkHighlight, ok := f[i]
		if !ok {
			f[i] = newHunkHighlight
			continue
		}

		f[i] = oldHunkHighlight.Merge(newHunkHighlight)
	}
	return f
}

type FileHighlight struct {
	OldFile Ranges
	NewFile Ranges
	Hunk    HunkHighlights
}

func (f FileHighlight) Merge(other FileHighlight) FileHighlight {
	f.OldFile = append(f.OldFile, other.OldFile...)
	f.NewFile = append(f.NewFile, other.NewFile...)
	f.Hunk = f.Hunk.Merge(other.Hunk)
	return f
}

type HunkHighlights map[int]LineHighlights

func (h HunkHighlights) Merge(other HunkHighlights) HunkHighlights {
	if len(other) == 0 {
		return h
	}

	if len(h) == 0 {
		return other
	}

	for i, newLineHighlight := range other {
		oldLineHighlight, ok := h[i]
		if !ok {
			h[i] = newLineHighlight
			continue
		}

		h[i] = oldLineHighlight.Merge(newLineHighlight)
	}
	return h
}

type LineHighlights map[int]Ranges

func (l LineHighlights) Merge(other LineHighlights) LineHighlights {
	if len(other) == 0 {
		return l
	}

	if len(l) == 0 {
		return other
	}

	for i, ranges := range other {
		oldRanges, ok := l[i]
		if !ok {
			l[i] = ranges
			continue
		}

		combined := append(oldRanges, ranges...)
		combined.MergeOverlapping()
		l[i] = combined
	}
	return l
}

type Location struct {
	Offset int `json:"offset"`
	// TODO add line and column as well
	// Line   int `json:"line"`
	// Column int `json:"column"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Ranges []Range

func (r Ranges) Len() int           { return len(r) }
func (r Ranges) Less(i, j int) bool { return r[i].Start.Offset < r[j].Start.Offset }
func (r Ranges) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

// MergeOverlapping simplifies a set of ranges by merging overlapping ranges
// into longer ranges. The reusulting set of ranges is guaranteed to be ordered
// by start offset and non-overlapping.
// TODO test this, especially around out-of-bounds conditions
func (r Ranges) MergeOverlapping() {
	sort.Sort(r)

	for i := 0; i+1 < len(r); {
		a := r[i]
		b := r[i+1]

		if b.Start.Offset <= a.End.Offset {
			r[i] = Range{
				Start: a.Start,
				End:   b.End,
			}

			// slide remaining elements left
			copy(r[i+1:], r[i+2:])
			r = r[:len(r)-1]
			continue
		}

		i++
	}
}
