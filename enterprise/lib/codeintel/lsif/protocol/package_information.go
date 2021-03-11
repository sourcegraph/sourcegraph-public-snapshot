package protocol

type PackageInformation struct {
	Vertex
	Name    string `json:"name"`
	Manager string `json:"manager"`
	Version string `json:"version"`
}

func NewPackageInformation(id uint64, name, manager, version string) PackageInformation {
	return PackageInformation{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexPackageInformation,
		},
		Name:    name,
		Manager: manager,
		Version: version,
	}
}

type PackageInformationEdge struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewPackageInformationEdge(id, outV, inV uint64) PackageInformationEdge {
	return PackageInformationEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgePackageInformation,
		},
		OutV: outV,
		InV:  inV,
	}
}
