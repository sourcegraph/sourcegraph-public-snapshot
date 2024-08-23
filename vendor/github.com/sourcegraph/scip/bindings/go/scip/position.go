package scip

import "fmt"

// Range represents [start, end) between two offset positions.
//
// NOTE: the github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol package
// contains similarly shaped structs but this one exists primarily to make it
// easier to work with SCIP encoded positions, which have the type []int32
// in Protobuf payloads.
type Range struct {
	Start Position
	End   Position
}

// Position represents an offset position.
type Position struct {
	Line      int32
	Character int32
}

func (p Position) Compare(other Position) int {
	if p.Line < other.Line {
		return -1
	}
	if p.Line > other.Line {
		return 1
	}
	if p.Character < other.Character {
		return -1
	}
	if p.Character > other.Character {
		return 1
	}
	return 0
}

func (p Position) Less(other Position) bool {
	if p.Line < other.Line {
		return true
	}
	if p.Line > other.Line {
		return false
	}
	return p.Character < other.Character
}

//go:noinline
func makeNewRangeError(startLine, endLine, startChar, endChar int32) (Range, error) {
	if startLine < 0 || endLine < 0 || startChar < 0 || endChar < 0 {
		return Range{}, NegativeOffsetsRangeError
	}
	if startLine > endLine || (startLine == endLine && startChar > endChar) {
		return Range{}, EndBeforeStartRangeError
	}
	panic("unreachable")
}

// NewRange constructs a Range while checking if the input is valid.
func NewRange(scipRange []int32) (Range, error) {
	// N.B. This function is kept small so that it can be inlined easily.
	// See also: https://github.com/golang/go/issues/17566
	var startLine, endLine, startChar, endChar int32
	switch len(scipRange) {
	case 3:
		startLine = scipRange[0]
		endLine = startLine
		startChar = scipRange[1]
		endChar = scipRange[2]
		if startLine >= 0 && startChar >= 0 && endChar >= startChar {
			break
		}
		return makeNewRangeError(startLine, endLine, startChar, endChar)
	case 4:
		startLine = scipRange[0]
		startChar = scipRange[1]
		endLine = scipRange[2]
		endChar = scipRange[3]
		if startLine >= 0 && startChar >= 0 &&
			((endLine > startLine && endChar >= 0) || (endLine == startLine && endChar >= startChar)) {
			break
		}
		return makeNewRangeError(startLine, endLine, startChar, endChar)
	default:
		return Range{}, IncorrectLengthRangeError
	}
	return Range{Start: Position{Line: startLine, Character: startChar}, End: Position{Line: endLine, Character: endChar}}, nil
}

type RangeError int32

const (
	IncorrectLengthRangeError RangeError = iota
	NegativeOffsetsRangeError
	EndBeforeStartRangeError
)

var _ error = RangeError(0)

func (e RangeError) Error() string {
	switch e {
	case IncorrectLengthRangeError:
		return "incorrect length"
	case NegativeOffsetsRangeError:
		return "negative offsets"
	case EndBeforeStartRangeError:
		return "end before start"
	}
	panic("unhandled range error")
}

// NewRangeUnchecked converts an SCIP range into `Range`
//
// Pre-condition: The input slice must follow the SCIP range encoding.
// https://sourcegraph.com/github.com/sourcegraph/scip/-/blob/scip.proto?L646:18-646:23
func NewRangeUnchecked(scipRange []int32) Range {
	// Single-line case is most common
	endCharacter := scipRange[2]
	endLine := scipRange[0]
	if len(scipRange) == 4 { // multi-line
		endCharacter = scipRange[3]
		endLine = scipRange[2]
	}
	return Range{
		Start: Position{
			Line:      scipRange[0],
			Character: scipRange[1],
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

func (r Range) IsSingleLine() bool {
	return r.Start.Line == r.End.Line
}

func (r Range) SCIPRange() []int32 {
	if r.Start.Line == r.End.Line {
		return []int32{r.Start.Line, r.Start.Character, r.End.Character}
	}
	return []int32{r.Start.Line, r.Start.Character, r.End.Line, r.End.Character}
}

// Contains checks if position is within the range
func (r Range) Contains(position Position) bool {
	return !position.Less(r.Start) && position.Less(r.End)
}

// Intersects checks if two ranges intersect.
//
// case 1: r1.Start >= other.Start && r1.Start < other.End
// case 2: r2.Start >= r1.Start && r2.Start < r1.End
func (r Range) Intersects(other Range) bool {
	return r.Start.Less(other.End) && other.Start.Less(r.End)
}

// Compare compares two ranges.
//
// Returns 0 if the ranges intersect (not just if they're equal).
func (r Range) Compare(other Range) int {
	if r.Intersects(other) {
		return 0
	}
	return r.Start.Compare(other.Start)
}

// Less compares two ranges, consistent with Compare.
//
// r.Compare(other) < 0 iff r.Less(other).
func (r Range) Less(other Range) bool {
	return r.End.Compare(other.Start) <= 0
}

// CompareStrict compares two ranges.
//
// Returns 0 iff the ranges are exactly equal.
func (r Range) CompareStrict(other Range) int {
	if ret := r.Start.Compare(other.Start); ret != 0 {
		return ret
	}
	return r.End.Compare(other.End)
}

// LessStrict compares two ranges, consistent with CompareStrict.
//
// r.CompareStrict(other) < 0 iff r.LessStrict(other).
func (r Range) LessStrict(other Range) bool {
	if ret := r.Start.Compare(other.Start); ret != 0 {
		return ret < 0
	}
	return r.End.Less(other.End)
}

func (r Range) String() string {
	return fmt.Sprintf("%d:%d-%d:%d", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
}
