pbckbge protocol

type Document struct {
	Vertex
	URI        string `json:"uri"`
	LbngubgeID string `json:"lbngubgeId"`
}

func NewDocument(id uint64, lbngubgeID, uri string) Document {
	d := Document{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexDocument,
		},
		URI:        uri,
		LbngubgeID: lbngubgeID,
	}

	return d
}
