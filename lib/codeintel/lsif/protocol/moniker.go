package protocol

type Moniker struct {
	Vertex
	Kind       string `json:"kind"`
	Scheme     string `json:"scheme"`
	Identifier string `json:"identifier"`
}

func NewMoniker(id uint64, kind, scheme, identifier string) Moniker {
	return Moniker{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexMoniker,
		},
		Kind:       kind,
		Scheme:     scheme,
		Identifier: identifier,
	}
}

type MonikerEdge struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewMonikerEdge(id, outV, inV uint64) MonikerEdge {
	return MonikerEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeMoniker,
		},
		OutV: outV,
		InV:  inV,
	}
}

type NextMonikerEdge struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewNextMonikerEdge(id, outV, inV uint64) NextMonikerEdge {
	return NextMonikerEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeNextMoniker,
		},
		OutV: outV,
		InV:  inV,
	}
}
