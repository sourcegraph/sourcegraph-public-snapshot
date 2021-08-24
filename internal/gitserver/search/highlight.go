package search

import (
	"sort"
)

type CommitHighlights struct {
	Diff    DeltaHighlights
	Message Ranges
	// TODO we could potentially return highlights for author and committer as well
	// Author    Ranges
	// Committer Ranges
}

func (c *CommitHighlights) Merge(other *CommitHighlights) *CommitHighlights {
	if c == nil {
		return other
	}

	if other == nil {
		return c
	}

	c.Diff = c.Diff.Merge(other.Diff)
	c.Message = c.Message.Merge(other.Message)
	return c
}

type DeltaHighlight struct {
	Index             int
	OldFileHighlights Ranges
	NewFileHighlights Ranges
	Hunks             HunkHighlights
}

func (f DeltaHighlight) Merge(other DeltaHighlight) DeltaHighlight {
	f.OldFileHighlights = f.OldFileHighlights.Merge(other.OldFileHighlights)
	f.NewFileHighlights = f.NewFileHighlights.Merge(other.NewFileHighlights)
	f.Hunks = f.Hunks.Merge(other.Hunks)
	return f
}

type DeltaHighlights []DeltaHighlight

func (f DeltaHighlights) Len() int           { return len(f) }
func (f DeltaHighlights) Less(i, j int) bool { return f[i].Index < f[j].Index }
func (f DeltaHighlights) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f DeltaHighlights) Merge(other DeltaHighlights) DeltaHighlights {
	f = append(f, other...)
	sort.Sort(f)

	unique := 0
	for i := 1; i < len(f); i++ {
		if f[unique].Index != f[i].Index {
			unique++
			f[unique] = f[i]
			continue
		}

		f[unique] = f[unique].Merge(f[i])
	}
	return f
}

type HunkHighlight struct {
	Index int
	Lines LineHighlights
}

func (h HunkHighlight) Merge(other HunkHighlight) HunkHighlight {
	h.Lines = h.Lines.Merge(other.Lines)
	return h
}

type HunkHighlights []HunkHighlight

func (h HunkHighlights) Len() int           { return len(h) }
func (h HunkHighlights) Less(i, j int) bool { return h[i].Index < h[j].Index }
func (h HunkHighlights) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h HunkHighlights) Merge(other HunkHighlights) HunkHighlights {
	h = append(h, other...)
	sort.Sort(h)

	unique := 0
	for i := 1; i < len(h); i++ {
		if h[unique].Index != h[i].Index {
			unique++
			h[unique] = h[i]
			continue
		}

		h[unique] = h[unique].Merge(h[i])
	}

	return h
}

type LineHighlight struct {
	Index      int
	Highlights Ranges
}

func (l LineHighlight) Merge(other LineHighlight) LineHighlight {
	l.Highlights = l.Highlights.Merge(other.Highlights)
	return l
}

type LineHighlights []LineHighlight

func (l LineHighlights) Len() int           { return len(l) }
func (l LineHighlights) Less(i, j int) bool { return l[i].Index < l[j].Index }
func (l LineHighlights) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l LineHighlights) Merge(other LineHighlights) LineHighlights {
	l = append(l, other...)
	sort.Sort(l)

	unique := 0
	for i := 1; i < len(l); i++ {
		if l[unique].Index != l[i].Index {
			unique++
			l[unique] = l[i]
			continue
		}

		l[unique] = l[unique].Merge(l[i])
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

func (r Ranges) Merge(other Ranges) Ranges {
	r = append(r, other...)
	sort.Sort(r)

	// Do not merge overlapping ranges because we want the result count to be accurate

	return r
}
