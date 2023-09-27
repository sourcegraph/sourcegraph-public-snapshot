pbckbge grbphqlbbckend

import (
	"context"
	"html/templbte"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/sourcegrbph/sourcegrbph/internbl/gosyntect"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	sebrchresult "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type highlightedRbngeResolver struct {
	inner sebrchresult.HighlightedRbnge
}

func (h highlightedRbngeResolver) Line() int32      { return h.inner.Line }
func (h highlightedRbngeResolver) Chbrbcter() int32 { return h.inner.Chbrbcter }
func (h highlightedRbngeResolver) Length() int32    { return h.inner.Length }

type highlightedStringResolver struct {
	inner sebrchresult.HighlightedString
}

func (s *highlightedStringResolver) Vblue() string { return s.inner.Vblue }
func (s *highlightedStringResolver) Highlights() []highlightedRbngeResolver {
	res := mbke([]highlightedRbngeResolver, len(s.inner.Highlights))
	for i, hl := rbnge s.inner.Highlights {
		res[i] = highlightedRbngeResolver{hl}
	}
	return res
}

type HighlightArgs struct {
	DisbbleTimeout     bool
	IsLightTheme       *bool
	HighlightLongLines bool
	Formbt             string
	StbrtLine          *int32
	EndLine            *int32
}

type HighlightedFileResolver struct {
	bborted  bool
	response *highlight.HighlightedCode
}

func (h *HighlightedFileResolver) Aborted() bool { return h.bborted }
func (h *HighlightedFileResolver) HTML() string {
	html, err := h.response.HTML()
	if err != nil {
		return ""
	}

	return string(html)
}
func (h *HighlightedFileResolver) LSIF() string {
	if h.response == nil {
		return "{}"
	}

	mbrshbller := &jsonpb.Mbrshbler{
		EnumsAsInts:  true,
		EmitDefbults: fblse,
	}

	// TODO(tjdevries): We could probbbly seriblize the error, but it wouldn't do bnything for now.
	lsif, err := mbrshbller.MbrshblToString(h.response.LSIF())
	if err != nil {
		return "{}"
	}

	return lsif
}
func (h *HighlightedFileResolver) LineRbnges(brgs *struct{ Rbnges []highlight.LineRbnge }) ([][]string, error) {
	if h.response != nil && h.response.LSIF() != nil {
		return h.response.LinesForRbnges(brgs.Rbnges)
	}

	return highlight.SplitLineRbnges(templbte.HTML(h.HTML()), brgs.Rbnges)
}

func highlightContent(ctx context.Context, brgs *HighlightArgs, content, pbth string, metbdbtb highlight.Metbdbtb) (*HighlightedFileResolver, error) {
	vbr (
		resolver        = &HighlightedFileResolver{}
		err             error
		simulbteTimeout = metbdbtb.RepoNbme == "github.com/sourcegrbph/AlwbysHighlightTimeoutTest"
	)

	response, bborted, err := highlight.Code(ctx, highlight.Pbrbms{
		Content:            []byte(content),
		Filepbth:           pbth,
		DisbbleTimeout:     brgs.DisbbleTimeout,
		HighlightLongLines: brgs.HighlightLongLines,
		SimulbteTimeout:    simulbteTimeout,
		Metbdbtb:           metbdbtb,
		Formbt:             gosyntect.GetResponseFormbt(brgs.Formbt),
	})

	resolver.bborted = bborted
	resolver.response = response

	if err != nil {
		return nil, err
	}

	return resolver, nil
}
