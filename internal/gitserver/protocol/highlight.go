package protocol

import (
	"encoding/json"
	"sort"
)

type CommitHighlights struct {
	Message Ranges
	Diff    map[int]FileDiffHighlight
}

func (c *CommitHighlights) Merge(other *CommitHighlights) *CommitHighlights {
	if c == nil {
		return other
	}

	if other == nil {
		return c
	}

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

type FileDiffHighlight struct {
	OldFile        Ranges
	NewFile        Ranges
	HunkHighlights map[int]HunkHighlight
}

func (f FileDiffHighlight) Merge(other FileDiffHighlight) FileDiffHighlight {
	f.OldFile = append(f.OldFile, other.OldFile...)
	sort.Sort(f.OldFile)

	f.NewFile = append(f.NewFile, other.NewFile...)
	sort.Sort(f.NewFile)

	if f.HunkHighlights == nil {
		f.HunkHighlights = other.HunkHighlights
	} else {
		for i, hh := range other.HunkHighlights {
			f.HunkHighlights[i] = f.HunkHighlights[i].Merge(hh)
		}
	}
	return f
}

type HunkHighlight struct {
	LineHighlights map[int]Ranges
}

func (h HunkHighlight) Merge(other HunkHighlight) HunkHighlight {
	if h.LineHighlights == nil {
		h.LineHighlights = other.LineHighlights
	} else {
		for i, lh := range other.LineHighlights {
			h.LineHighlights[i] = h.LineHighlights[i].Merge(lh)
		}
	}
	return h
}

type HighlightedString struct {
	Content    string `json:"content"`
	Highlights Ranges `json:"highlights"`
}

func (h *HighlightedString) Merge(other HighlightedString) {
	if h.Content == "" {
		h.Content = other.Content
	}
	h.Highlights = append(h.Highlights, other.Highlights...)
	// TODO(camdencheek): Do we need to guarantee that these are non-overlapping like Zoekt does?
	sort.Sort(h.Highlights)
}

type Location struct {
	Offset int
	Line   int
	Column int
}

func (l Location) Add(o Location) Location {
	return Location{
		Offset: l.Offset + o.Offset,
		Line:   l.Line + o.Line,
		Column: l.Column + o.Column,
	}
}

func (l Location) Sub(o Location) Location {
	return Location{
		Offset: l.Offset - o.Offset,
		Line:   l.Line - o.Line,
		Column: l.Column - o.Column,
	}
}

// MarshalJSON provides a custom JSON serialization to reduce
// the size overhead of sending the field names for every location
func (l Location) MarshalJSON() ([]byte, error) {
	return json.Marshal([3]int{l.Offset, l.Line, l.Column})
}

func (l *Location) UnmarshalJSON(data []byte) error {
	var v [3]int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	l.Offset = v[0]
	l.Line = v[1]
	l.Column = v[2]
	return nil
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

func (r Range) Includes(loc Location) bool {
	return r.Start.Offset <= loc.Offset && r.End.Offset > loc.Offset
}

func (r Range) Contains(r2 Range) bool {
	return r.Includes(r2.Start) && r.Includes(r2.End)
}

func (r Range) Add(amount Location) Range {
	return Range{
		Start: r.Start.Add(amount),
		End:   r.End.Add(amount),
	}
}

func (r Range) Sub(amount Location) Range {
	return Range{
		Start: r.Start.Sub(amount),
		End:   r.End.Sub(amount),
	}
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

func (r Ranges) Add(amount Location) Ranges {
	res := make(Ranges, 0, len(r))
	for _, oldRange := range r {
		res = append(res, oldRange.Add(amount))
	}
	return res
}

func (r Ranges) Sub(amount Location) Ranges {
	res := make(Ranges, 0, len(r))
	for _, oldRange := range r {
		res = append(res, oldRange.Sub(amount))
	}
	return res
}
