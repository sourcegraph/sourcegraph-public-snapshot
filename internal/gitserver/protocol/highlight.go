package protocol

import (
	"encoding/json"
	"sort"
)

type HighlightedCommit struct {
	Diff    HighlightedString
	Message HighlightedString

	// TODO we could potentially return highlights for author and committer as well
	// Author    Ranges
	// Committer Ranges
}

func (c *HighlightedCommit) Merge(other *HighlightedCommit) *HighlightedCommit {
	if c == nil {
		return other
	}

	if other == nil {
		return c
	}

	c.Diff.Merge(other.Diff)
	c.Message.Merge(other.Message)
	return c
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

func (l Location) Shift(o Location) Location {
	return Location{
		Offset: l.Offset + o.Offset,
		Line:   l.Line + o.Line,
		Column: l.Column + o.Column,
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

// Shift takes a set of ranges and creates a new set of ranges whose
// start and end locations are offset by the given amount. Note, we could
// mutate the ranges in place to avoid an allocation, but it's a relatively
// small cost for immutability.
func (r Ranges) Shift(amount Location) Ranges {
	res := make(Ranges, 0, len(r))
	for _, oldRange := range r {
		res = append(res, Range{
			Start: oldRange.Start.Shift(amount),
			End:   oldRange.End.Shift(amount),
		})
	}
	return res
}
