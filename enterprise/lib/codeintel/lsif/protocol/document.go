package protocol

type Document struct {
	Vertex
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
}

func NewDocument(id uint64, languageID, uri string) Document {
	d := Document{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexDocument,
		},
		URI:        uri,
		LanguageID: languageID,
	}

	return d
}
