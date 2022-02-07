package lsif_typed

import sitter "github.com/smacker/go-tree-sitter"

type RangePosition struct {
	Start Position
	End   Position
}

func (r RangePosition) IsSingleLine() bool {
	return r.Start.Line == r.End.Line
}
func (r RangePosition) LsifRange() []int32 {
	if r.Start.Line == r.End.Line {
		return []int32{int32(r.Start.Line), int32(r.Start.Character), int32(r.End.Character)}
	}
	return []int32{int32(r.Start.Line), int32(r.Start.Character), int32(r.End.Line), int32(r.End.Character)}
}

func NewRangePositionFromLsif(lsifRange []int32) *RangePosition {
	var endLine int32
	var endCharacter int32
	if len(lsifRange) == 3 { // single line
		endLine = lsifRange[0]
		endCharacter = lsifRange[2]
	} else if len(lsifRange) == 4 { // multi-line
		endLine = lsifRange[2]
		endCharacter = lsifRange[3]
	}
	return &RangePosition{
		Start: Position{
			Line:      int(lsifRange[0]),
			Character: int(lsifRange[1]),
		},
		End: Position{
			Line:      int(endLine),
			Character: int(endCharacter),
		},
	}
}

func NewRangePositionFromNode(node *sitter.Node) *RangePosition {
	return &RangePosition{
		Start: Position{
			Line:      int(node.StartPoint().Row),
			Character: int(node.StartPoint().Column),
		},
		End: Position{
			Line:      int(node.EndPoint().Row),
			Character: int(node.EndPoint().Column),
		},
	}
}
