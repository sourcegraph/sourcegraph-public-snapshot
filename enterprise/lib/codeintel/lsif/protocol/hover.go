package protocol

import jsoniter "github.com/json-iterator/go"

type HoverResult struct {
	Vertex
	Result hoverResult `json:"result"`
}

type hoverResult struct {
	Contents []MarkedString `json:"contents"`
}

func NewHoverResult(id uint64, contents []MarkedString) HoverResult {
	return HoverResult{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexHoverResult,
		},
		Result: hoverResult{
			Contents: contents,
		},
	}
}

type MarkedString markedString

type markedString struct {
	Language    string `json:"language"`
	Value       string `json:"value"`
	isRawString bool
}

func NewMarkedString(s, languageID string) MarkedString {
	return MarkedString{
		Language: languageID,
		Value:    s,
	}
}

func RawMarkedString(s string) MarkedString {
	return MarkedString{
		Value:       s,
		isRawString: true,
	}
}

var marshaller = jsoniter.ConfigFastest

func (m MarkedString) MarshalJSON() ([]byte, error) {
	if m.isRawString {
		return marshaller.Marshal(m.Value)
	}
	return marshaller.Marshal((markedString)(m))
}

type TextDocumentHover struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentHover(id, outV, inV uint64) TextDocumentHover {
	return TextDocumentHover{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeTextDocumentHover,
		},
		OutV: outV,
		InV:  inV,
	}
}
