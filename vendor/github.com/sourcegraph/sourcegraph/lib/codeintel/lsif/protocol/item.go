package protocol

type Item struct {
	Edge
	OutV     uint64   `json:"outV"`
	InVs     []uint64 `json:"inVs"`
	Document uint64   `json:"document"`
	Property string   `json:"property,omitempty"`
}

func NewItem(id, outV uint64, inVs []uint64, document uint64) Item {
	return Item{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeItem,
		},
		OutV:     outV,
		InVs:     inVs,
		Document: document,
	}
}

func NewItemWithProperty(id, outV uint64, inVs []uint64, document uint64, property string) Item {
	i := NewItem(id, outV, inVs, document)
	i.Property = property
	return i
}

func NewItemOfDefinitions(id, outV uint64, inVs []uint64, document uint64) Item {
	return NewItemWithProperty(id, outV, inVs, document, "definitions")
}

func NewItemOfReferences(id, outV uint64, inVs []uint64, document uint64) Item {
	return NewItemWithProperty(id, outV, inVs, document, "references")
}
