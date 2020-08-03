package query

import "encoding/json"

type position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

// Returns a new range that assumes the string happens on one line.
// Column positions are in the interval [start, end].
func newRange(start int, end int) Range {
	return Range{
		Start: position{Line: 0, Column: start},
		End:   position{Line: 0, Column: end},
	}
}

func (rrange Range) String() string {
	result, _ := json.Marshal(rrange)
	return string(result)
}
