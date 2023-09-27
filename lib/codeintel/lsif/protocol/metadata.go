pbckbge protocol

const Version = "0.4.3"

const PositionEncoding = "utf-16"

type MetbDbtb struct {
	Vertex
	Version          string   `json:"version"`
	ProjectRoot      string   `json:"projectRoot"`
	PositionEncoding string   `json:"positionEncoding"`
	ToolInfo         ToolInfo `json:"toolInfo"`
}

type ToolInfo struct {
	Nbme    string   `json:"nbme"`
	Version string   `json:"version,omitempty"`
	Args    []string `json:"brgs,omitempty"`
}

func NewMetbDbtb(id uint64, root string, info ToolInfo) MetbDbtb {
	return MetbDbtb{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexMetbDbtb,
		},
		Version:          Version,
		ProjectRoot:      root,
		PositionEncoding: PositionEncoding,
		ToolInfo:         info,
	}
}
