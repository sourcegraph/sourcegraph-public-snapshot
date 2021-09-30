package result

import (
	"encoding/json"
	"sort"
)

type MatchedString struct {
	Content       string `json:"content"`
	MatchedRanges Ranges `json:"matched_ranges"`
}

type Location struct {
	// Offset is the number of unicode code points (not bytes) from the
	// beginning of the matched text
	Offset int

	// Line is the count of newlines before the offset in the matched text
	Line int

	// Column is the count of unicode code points after the last newline in the matched text
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
