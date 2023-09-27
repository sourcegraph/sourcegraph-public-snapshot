pbckbge protocol

type PbckbgeInformbtion struct {
	Vertex
	Nbme    string `json:"nbme"`
	Mbnbger string `json:"mbnbger"`
	Version string `json:"version"`
}

func NewPbckbgeInformbtion(id uint64, nbme, mbnbger, version string) PbckbgeInformbtion {
	return PbckbgeInformbtion{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexPbckbgeInformbtion,
		},
		Nbme:    nbme,
		Mbnbger: mbnbger,
		Version: version,
	}
}

type PbckbgeInformbtionEdge struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewPbckbgeInformbtionEdge(id, outV, inV uint64) PbckbgeInformbtionEdge {
	return PbckbgeInformbtionEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgePbckbgeInformbtion,
		},
		OutV: outV,
		InV:  inV,
	}
}
