package protocol

type Range struct {
	Vertex
	RangeData
	Tag *RangeTag `json:"tag,omitempty"`
}

type RangeData struct {
	Start Pos `json:"start"`
	End   Pos `json:"end"`
}

// RangeTag represents a tag associated with a range that provides metadata about the symbol defined
// at the range. Some of the fields may be empty depending on the value of Type. See
// https://microsoft.github.io/language-server-protocol/specifications/lsif/0.4.0/specification/#documentSymbol
type RangeTag struct {
	Type      string     `json:"type"`
	Text      string     `json:"text"`
	Kind      SymbolKind `json:"kind"`
	FullRange *RangeData `json:"fullRange,omitempty"`
	Detail    string     `json:"detail,omitempty"`

	// Tags is a custom extension, see https://github.com/microsoft/language-server-protocol/issues/1209
	Tags []SymbolTag `json:"tags,omitempty"`
}

type Pos struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func NewRange(id uint64, start, end Pos, tag *RangeTag) Range {
	return Range{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexRange,
		},
		RangeData: RangeData{
			Start: start,
			End:   end,
		},
		Tag: tag,
	}
}
