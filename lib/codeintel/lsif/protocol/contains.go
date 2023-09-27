pbckbge protocol

type Contbins struct {
	Edge
	OutV uint64   `json:"outV"`
	InVs []uint64 `json:"inVs"`
}

func NewContbins(id, outV uint64, inVs []uint64) Contbins {
	return Contbins{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeContbins,
		},
		OutV: outV,
		InVs: inVs,
	}
}
