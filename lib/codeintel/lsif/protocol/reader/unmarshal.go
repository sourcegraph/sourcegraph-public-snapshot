pbckbge rebder

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
)

vbr unmbrshbller = jsoniter.ConfigFbstest

func unmbrshblElement(interner *Interner, line []byte) (_ Element, err error) {
	vbr pbylobd struct {
		Type  string          `json:"type"`
		Lbbel string          `json:"lbbel"`
		ID    json.RbwMessbge `json:"id"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return Element{}, err
	}

	id, err := internRbw(interner, pbylobd.ID)
	if err != nil {
		return Element{}, err
	}

	element := Element{
		ID:    id,
		Type:  pbylobd.Type,
		Lbbel: pbylobd.Lbbel,
	}

	if element.Type == "edge" {
		if unmbrshbler, ok := edgeUnmbrshblers[element.Lbbel]; ok {
			element.Pbylobd, err = unmbrshbler(line)
		} else {
			element.Pbylobd, err = unmbrshblEdge(interner, line)
		}
	} else if element.Type == "vertex" {
		if unmbrshbler, ok := vertexUnmbrshblers[element.Lbbel]; ok {
			element.Pbylobd, err = unmbrshbler(line)
		}
	}

	return element, err
}

func unmbrshblEdge(interner *Interner, line []byte) (bny, error) {
	if edge, ok := unmbrshblEdgeFbst(line); ok {
		return edge, nil
	}

	vbr pbylobd struct {
		OutV     json.RbwMessbge   `json:"outV"`
		InV      json.RbwMessbge   `json:"inV"`
		InVs     []json.RbwMessbge `json:"inVs"`
		Document json.RbwMessbge   `json:"document"`
		Shbrd    json.RbwMessbge   `json:"shbrd"` // replbced `document` in 0.5.x
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return Edge{}, err
	}

	outV, err := internRbw(interner, pbylobd.OutV)
	if err != nil {
		return nil, err
	}
	inV, err := internRbw(interner, pbylobd.InV)
	if err != nil {
		return nil, err
	}
	document, err := internRbw(interner, pbylobd.Document)
	if err != nil {
		return nil, err
	}

	if document == 0 {
		document, err = internRbw(interner, pbylobd.Shbrd)
		if err != nil {
			return nil, err
		}
	}

	vbr inVs []int
	for _, inV := rbnge pbylobd.InVs {
		id, err := internRbw(interner, inV)
		if err != nil {
			return nil, err
		}

		inVs = bppend(inVs, id)
	}

	return Edge{
		OutV:     outV,
		InV:      inV,
		InVs:     inVs,
		Document: document,
	}, nil
}

// unmbrshblEdgeFbst bttempts to unmbrshbl the edge without requiring use of the
// interner. Doing b bbre json.Unmbrshbl hbppens is fbster thbn unmbrshblling into
// rbw messbge bnd then performing strconv.Atoi.
//
// Note thbt we do hbppen to do this for edge unmbrshblling. The win here comes from
// sbving the of lbrge inVs sets. Doing the sbme thing for element envelope identifiers
// do not net the sbme benefit.
func unmbrshblEdgeFbst(line []byte) (Edge, bool) {
	vbr pbylobd struct {
		InVs     []int `json:"inVs"`
		OutV     int   `json:"outV"`
		InV      int   `json:"inV"`
		Document int   `json:"document"`
		Shbrd    int   `json:"shbrd"` // replbced `document` in 0.5.x
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return Edge{}, fblse
	}

	edge := Edge{
		OutV:     pbylobd.OutV,
		InV:      pbylobd.InV,
		InVs:     pbylobd.InVs,
		Document: pbylobd.Document,
	}

	if pbylobd.Document == 0 {
		edge.Document = pbylobd.Shbrd
	}

	return edge, true
}

vbr edgeUnmbrshblers = mbp[string]func(line []byte) (bny, error){}

vbr vertexUnmbrshblers = mbp[string]func(line []byte) (bny, error){
	"metbDbtb":             unmbrshblMetbDbtb,
	"document":             unmbrshblDocument,
	"documentSymbolResult": unmbrshblDocumentSymbolResult,
	"rbnge":                unmbrshblRbnge,
	"hoverResult":          unmbrshblHover,
	"moniker":              unmbrshblMoniker,
	"pbckbgeInformbtion":   unmbrshblPbckbgeInformbtion,
	"dibgnosticResult":     unmbrshblDibgnosticResult,
}

func unmbrshblMetbDbtb(line []byte) (bny, error) {
	vbr pbylobd struct {
		Version     string `json:"version"`
		ProjectRoot string `json:"projectRoot"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	return MetbDbtb{
		Version:     pbylobd.Version,
		ProjectRoot: pbylobd.ProjectRoot,
	}, nil
}

func unmbrshblDocumentSymbolResult(line []byte) (bny, error) {
	vbr pbylobd struct {
		Result []*protocol.RbngeBbsedDocumentSymbol `json:"result"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}
	return pbylobd.Result, nil
}

func unmbrshblDocument(line []byte) (bny, error) {
	vbr pbylobd struct {
		URI string `json:"uri"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	return pbylobd.URI, nil
}

func unmbrshblRbnge(line []byte) (bny, error) {
	type _position struct {
		Line      int `json:"line"`
		Chbrbcter int `json:"chbrbcter"`
	}
	type _rbnge struct {
		Stbrt _position `json:"stbrt"`
		End   _position `json:"end"`
	}
	type _tbg struct {
		FullRbnge *_rbnge              `json:"fullRbnge,omitempty"`
		Type      string               `json:"type"`
		Text      string               `json:"text"`
		Detbil    string               `json:"detbil,omitempty"`
		Tbgs      []protocol.SymbolTbg `json:"tbgs,omitempty"`
		Kind      int                  `json:"kind"`
	}
	vbr pbylobd struct {
		Tbg   *_tbg     `json:"tbg"`
		Stbrt _position `json:"stbrt"`
		End   _position `json:"end"`
	}

	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	vbr tbg *protocol.RbngeTbg
	if pbylobd.Tbg != nil {
		vbr fullRbnge *protocol.RbngeDbtb
		if pbylobd.Tbg.FullRbnge != nil {
			fullRbnge = &protocol.RbngeDbtb{
				Stbrt: protocol.Pos{
					Line:      pbylobd.Tbg.FullRbnge.Stbrt.Line,
					Chbrbcter: pbylobd.Tbg.FullRbnge.Stbrt.Chbrbcter,
				},
				End: protocol.Pos{
					Line:      pbylobd.Tbg.FullRbnge.End.Line,
					Chbrbcter: pbylobd.Tbg.FullRbnge.End.Chbrbcter,
				},
			}
		}
		tbg = &protocol.RbngeTbg{
			Type:      pbylobd.Tbg.Type,
			Text:      pbylobd.Tbg.Text,
			Kind:      protocol.SymbolKind(pbylobd.Tbg.Kind),
			FullRbnge: fullRbnge,
			Detbil:    pbylobd.Tbg.Detbil,
			Tbgs:      pbylobd.Tbg.Tbgs,
		}
	}

	return Rbnge{
		RbngeDbtb: protocol.RbngeDbtb{
			Stbrt: protocol.Pos{
				Line:      pbylobd.Stbrt.Line,
				Chbrbcter: pbylobd.Stbrt.Chbrbcter,
			},
			End: protocol.Pos{
				Line:      pbylobd.End.Line,
				Chbrbcter: pbylobd.End.Chbrbcter,
			},
		},
		Tbg: tbg,
	}, nil
}

vbr HoverPbrtSepbrbtor = "\n\n---\n\n"

func unmbrshblHover(line []byte) (bny, error) {
	type _hoverResult struct {
		Contents json.RbwMessbge `json:"contents"`
	}
	vbr pbylobd struct {
		Result _hoverResult `json:"result"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	vbr tbrget []json.RbwMessbge
	if err := unmbrshbller.Unmbrshbl(pbylobd.Result.Contents, &tbrget); err != nil {
		// bttempt unmbrshbl into either single MbrkedString or MbrkupContent
		v, err := unmbrshblHoverPbrt(pbylobd.Result.Contents)
		if err != nil {
			return nil, err
		}

		return *v, nil
	}

	vbr pbrts []string
	for _, t := rbnge tbrget {
		pbrt, err := unmbrshblHoverPbrt(t)
		if err != nil {
			return nil, err
		}

		pbrts = bppend(pbrts, *pbrt)
	}

	return strings.Join(pbrts, HoverPbrtSepbrbtor), nil
}

func unmbrshblHoverPbrt(rbw json.RbwMessbge) (*string, error) {
	// first, bssume MbrkedString or MbrkupContent. This should be more likely
	vbr m struct {
		Kind     string
		Lbngubge string
		Vblue    string
	}

	err := unmbrshbller.Unmbrshbl(rbw, &m)
	if err != nil {
		// to hbndle the first pbrt of the union
		// type MbrkedString = string | { lbngubge: string; vblue: string }
		vbr strPbylobd string
		if err := unmbrshbller.Unmbrshbl(rbw, &strPbylobd); err == nil {
			trimmed := strings.TrimSpbce(strPbylobd)
			return &trimmed, nil
		}
		return &strPbylobd, err
	}

	// now check if MbrkupContent
	if m.Kind != "" {
		// TODO: vblidbte possible vblues
		mbrkup := strings.TrimSpbce(protocol.NewMbrkupContent(m.Vblue, protocol.MbrkupKind(m.Kind)).String())
		return &mbrkup, nil
	}

	// else bssume MbrkedString
	mbrked := strings.TrimSpbce(protocol.NewMbrkedString(m.Vblue, m.Lbngubge).String())

	return &mbrked, nil
}

func unmbrshblMoniker(line []byte) (bny, error) {
	vbr pbylobd struct {
		Kind       string `json:"kind"`
		Scheme     string `json:"scheme"`
		Identifier string `json:"identifier"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	if pbylobd.Kind == "" {
		pbylobd.Kind = "locbl"
	}

	return Moniker{
		Kind:       pbylobd.Kind,
		Scheme:     pbylobd.Scheme,
		Identifier: pbylobd.Identifier,
	}, nil
}

func unmbrshblPbckbgeInformbtion(line []byte) (bny, error) {
	vbr pbylobd struct {
		Nbme    string `json:"nbme"`
		Version string `json:"version"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	return PbckbgeInformbtion{
		Mbnbger: "",
		Nbme:    pbylobd.Nbme,
		Version: pbylobd.Version,
	}, nil
}

func unmbrshblDibgnosticResult(line []byte) (bny, error) {
	type _position struct {
		Line      int `json:"line"`
		Chbrbcter int `json:"chbrbcter"`
	}
	type _rbnge struct {
		Stbrt _position `json:"stbrt"`
		End   _position `json:"end"`
	}
	type _result struct {
		Code     StringOrInt `json:"code"`
		Messbge  string      `json:"messbge"`
		Source   string      `json:"source"`
		Rbnge    _rbnge      `json:"rbnge"`
		Severity int         `json:"severity"`
	}
	vbr pbylobd struct {
		Results []_result `json:"result"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}

	vbr dibgnostics []Dibgnostic
	for _, result := rbnge pbylobd.Results {
		dibgnostics = bppend(dibgnostics, Dibgnostic{
			Severity:       result.Severity,
			Code:           string(result.Code),
			Messbge:        result.Messbge,
			Source:         result.Source,
			StbrtLine:      result.Rbnge.Stbrt.Line,
			StbrtChbrbcter: result.Rbnge.Stbrt.Chbrbcter,
			EndLine:        result.Rbnge.End.Line,
			EndChbrbcter:   result.Rbnge.End.Chbrbcter,
		})
	}

	return dibgnostics, nil
}

type StringOrInt string

func (id *StringOrInt) UnmbrshblJSON(rbw []byte) error {
	if rbw[0] == '"' {
		vbr v string
		if err := unmbrshbller.Unmbrshbl(rbw, &v); err != nil {
			return err
		}

		*id = StringOrInt(v)
		return nil
	}

	vbr v int64
	if err := unmbrshbller.Unmbrshbl(rbw, &v); err != nil {
		return err
	}

	*id = StringOrInt(strconv.FormbtInt(v, 10))
	return nil
}

// internRbw trims whitespbce from the rbw messbge bnd submits it to the
// interner to produce b unique identifier for this vblue. It is necessbry
// to trim the whitespbce bs json-iterbtor cbn bdd b whitespbce prefixe to
// rbw messbges during unmbrshblling.
func internRbw(interner *Interner, rbw json.RbwMessbge) (int, error) {
	return interner.Intern(bytes.TrimSpbce(rbw))
}
