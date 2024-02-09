package result

import (
	"bufio"
	"cmp"
	"encoding/json"
	"sort"
	"strings"
)

type MatchedString struct {
	Content       string `json:"content"`
	MatchedRanges Ranges `json:"matchedRanges"`
}

func (m MatchedString) ToHighlightedString() HighlightedString {
	highlights := make([]HighlightedRange, 0, len(m.MatchedRanges))
	for _, r := range m.MatchedRanges {
		highlights = append(highlights, rangeToHighlights(m.Content, r)...)
	}
	return HighlightedString{Value: m.Content, Highlights: highlights}
}

// rangeToHighlights converts a Range (which can cross multiple lines)
// into HighlightedRange, which is scoped to one line. In order to do this
// correctly, we need the string that is being highlighted in order to identify
// line-end boundaries within multi-line ranges.
// TODO(camdencheek): push the Range format up the stack so we can be smarter about multi-line highlights.
func rangeToHighlights(s string, r Range) []HighlightedRange {
	var res []HighlightedRange

	// Use a scanner to handle \r?\n
	scanner := bufio.NewScanner(strings.NewReader(s[r.Start.Offset:r.End.Offset]))
	lineNum := r.Start.Line
	for scanner.Scan() {
		line := scanner.Text()

		character := 0
		if lineNum == r.Start.Line {
			character = r.Start.Column
		}

		length := len(line)
		if lineNum == r.End.Line {
			length = r.End.Column - character
		}

		if length > 0 {
			res = append(res, HighlightedRange{
				Line:      int32(lineNum),
				Character: int32(character),
				Length:    int32(length),
			})
		}

		lineNum++
	}

	return res
}

// Location represents the location of a character in some UTF-8 encoded content.
type Location struct {
	// Offset is the number of bytes preceding this character in the content
	Offset int

	// Line is the count of newlines before the offset in the matched text
	Line int

	// Column is the count of UTF-8 runes after the last newline in the matched text
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

// Compare compares the Offset of l and o.
func (l Location) Compare(o Location) int {
	return cmp.Compare(l.Offset, o.Offset)
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

// Range represents a slice [start, end) of some UTF-8 encoded content.
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
