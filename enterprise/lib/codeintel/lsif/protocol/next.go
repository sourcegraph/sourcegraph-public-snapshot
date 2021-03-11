package protocol

type Next struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewNext(id, outV, inV uint64) Next {
	return Next{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeNext,
		},
		OutV: outV,
		InV:  inV,
	}
}
