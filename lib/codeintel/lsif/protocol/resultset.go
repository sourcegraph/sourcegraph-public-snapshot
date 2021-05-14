package protocol

type ResultSet struct {
	Vertex
}

func NewResultSet(id uint64) ResultSet {
	return ResultSet{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexResultSet,
		},
	}
}
