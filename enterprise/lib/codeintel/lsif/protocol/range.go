package protocol

type Range struct {
	Vertex
	Start Pos `json:"start"`
	End   Pos `json:"end"`
}

type Pos struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func NewRange(id uint64, start, end Pos) Range {
	return Range{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexRange,
		},
		Start: start,
		End:   end,
	}
}
