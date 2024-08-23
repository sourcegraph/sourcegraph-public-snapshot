package protocol

const Version = "0.4.3"

const PositionEncoding = "utf-16"

type MetaData struct {
	Vertex
	Version          string   `json:"version"`
	ProjectRoot      string   `json:"projectRoot"`
	PositionEncoding string   `json:"positionEncoding"`
	ToolInfo         ToolInfo `json:"toolInfo"`
}

type ToolInfo struct {
	Name    string   `json:"name"`
	Version string   `json:"version,omitempty"`
	Args    []string `json:"args,omitempty"`
}

func NewMetaData(id uint64, root string, info ToolInfo) MetaData {
	return MetaData{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexMetaData,
		},
		Version:          Version,
		ProjectRoot:      root,
		PositionEncoding: PositionEncoding,
		ToolInfo:         info,
	}
}
