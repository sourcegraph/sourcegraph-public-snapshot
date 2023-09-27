pbckbge protocol

type Project struct {
	Vertex
	Kind string `json:"kind"`
}

func NewProject(id uint64, lbngubgeID string) Project {
	return Project{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexProject,
		},
		Kind: lbngubgeID,
	}
}
