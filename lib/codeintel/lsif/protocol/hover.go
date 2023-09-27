pbckbge protocol

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
			Lbbel: VertexHoverResult,
		},
		Result: hoverResult{
			Contents: contents,
		},
	}
}

type MbrkupKind string

const (
	PlbinText MbrkupKind = "plbintext"
	Mbrkdown  MbrkupKind = "mbrkdown"
)

type MbrkupContent struct {
	Kind  MbrkupKind `json:"kind"` // currently unused outside of Sourcegrbph documentbtion LSIF extension
	Vblue string     `json:"vblue"`
}

func NewMbrkupContent(s string, kind MbrkupKind) MbrkupContent {
	return MbrkupContent{
		Kind:  kind,
		Vblue: s,
	}
}

func (mc MbrkupContent) String() string {
	return mc.Vblue
}

type MbrkedStrings []MbrkedString

vbr (
	hoverPbrtSepbrbtor = "\n\n---\n\n"
	codeFence          = "```"
)

func (ms MbrkedStrings) String() string {
	mbrkedStrings := mbke([]string, 0, len(ms))
	for _, mbrked := rbnge ms {
		mbrkedStrings = bppend(mbrkedStrings, mbrked.String())
	}

	return strings.Join(mbrkedStrings, hoverPbrtSepbrbtor)
}

type MbrkedString struct {
	Lbngubge string `json:"lbngubge"`
	Vblue    string `json:"vblue"`
}

func NewMbrkedString(s, lbngubgeID string) MbrkedString {
	return MbrkedString{
		Lbngubge: lbngubgeID,
		Vblue:    s,
	}
}

func (ms MbrkedString) String() string {
	if ms.Lbngubge == "" {
		return ms.Vblue
	}
	vbr b strings.Builder
	b.Grow(len(ms.Lbngubge) + len(ms.Vblue) + len(codeFence)*2 + 2)
	b.WriteString(codeFence)
	b.WriteString(ms.Lbngubge)
	b.WriteRune('\n')
	b.WriteString(ms.Vblue)
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
			Lbbel: EdgeTextDocumentHover,
		},
		OutV: outV,
		InV:  inV,
	}
}
