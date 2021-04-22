package protocol

import (
	"fmt"
	"strings"
)

type HoverResult struct {
	Vertex
	Result hoverResult `json:"result"`
}

type hoverResult struct {
	Contents fmt.Stringer `json:"contents"`
}

func NewHoverResult(id uint64, contents fmt.Stringer) HoverResult {
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

type MarkupKind string

const (
	PlainText MarkupKind = "plaintext"
	Markdown  MarkupKind = "markdown"
)

type MarkupContent struct {
	Kind  MarkupKind `json:"kind"` // currently unused
	Value string     `json:"value"`
}

func NewMarkupContent(s string, kind MarkupKind) MarkupContent {
	return MarkupContent{
		Kind:  kind,
		Value: s,
	}
}

func (mc MarkupContent) String() string {
	return mc.Value
}

type MarkedStrings []MarkedString

var (
	hoverPartSeparator = "\n\n---\n\n"
	codeFence          = "```"
)

func (ms MarkedStrings) String() string {
	markedStrings := make([]string, 0, len(ms))
	for _, marked := range ms {
		markedStrings = append(markedStrings, marked.String())
	}

	return strings.Join(markedStrings, hoverPartSeparator)
}

type MarkedString struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

func NewMarkedString(s, languageID string) MarkedString {
	return MarkedString{
		Language: languageID,
		Value:    s,
	}
}

func (ms MarkedString) String() string {
	if ms.Language == "" {
		return ms.Value
	}
	var b strings.Builder
	b.Grow(len(ms.Language) + len(ms.Value) + len(codeFence)*2 + 2)
	b.WriteString(codeFence)
	b.WriteString(ms.Language)
	b.WriteRune('\n')
	b.WriteString(ms.Value)
	b.WriteRune('\n')
	b.WriteString(codeFence)

	return b.String()
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
