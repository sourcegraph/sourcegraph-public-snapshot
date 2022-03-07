package lsiftyped

// Range represents a range between two offset positions.
// NOTE: the lsif/protocol package contains similarly shaped structs but this
// one exists primarily to make it easier to work with LSIF Typed encoded positions,
// which have the type []int32 in Protobuf payloads.
type Range struct {
	Start Position
	End   Position
}

// Position represents an offset position.
type Position struct {
	Line      int32
	Character int32
}

// NewRange converts an LSIF Typed range into `Range`
func NewRange(lsifRange []int32) *Range {
	var endLine int32
	var endCharacter int32
	if len(lsifRange) == 3 { // single line
		endLine = lsifRange[0]
		endCharacter = lsifRange[2]
	} else if len(lsifRange) == 4 { // multi-line
		endLine = lsifRange[2]
		endCharacter = lsifRange[3]
	}
	return &Range{
		Start: Position{
			Line:      lsifRange[0],
			Character: lsifRange[1],
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

func (r Range) LsifRange() []int32 {
	if r.Start.Line == r.End.Line {
		return []int32{r.Start.Line, r.Start.Character, r.End.Character}
	}
	return []int32{r.Start.Line, r.Start.Character, r.End.Line, r.End.Character}
}
