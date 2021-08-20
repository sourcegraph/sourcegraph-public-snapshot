package protocol

type Contains struct {
	Edge
	OutV uint64   `json:"outV"`
	InVs []uint64 `json:"inVs"`
}

func NewContains(id, outV uint64, inVs []uint64) Contains {
	return Contains{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeContains,
		},
		OutV: outV,
		InVs: inVs,
	}
}
