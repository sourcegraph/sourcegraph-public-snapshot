package reader

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
)

var unmarshaller = jsoniter.ConfigFastest

func unmarshalElement(interner *Interner, line []byte) (_ Element, err error) {
	var payload struct {
		Type  string          `json:"type"`
		Label string          `json:"label"`
		ID    json.RawMessage `json:"id"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return Element{}, err
	}

	id, err := internRaw(interner, payload.ID)
	if err != nil {
		return Element{}, err
	}

	element := Element{
		ID:    id,
		Type:  payload.Type,
		Label: payload.Label,
	}

	if element.Type == "edge" {
		if unmarshaler, ok := edgeUnmarshalers[element.Label]; ok {
			element.Payload, err = unmarshaler(line)
		} else {
			element.Payload, err = unmarshalEdge(interner, line)
		}
	} else if element.Type == "vertex" {
		if unmarshaler, ok := vertexUnmarshalers[element.Label]; ok {
			element.Payload, err = unmarshaler(line)
		}
	}

	return element, err
}

func unmarshalEdge(interner *Interner, line []byte) (any, error) {
	if edge, ok := unmarshalEdgeFast(line); ok {
		return edge, nil
	}

	var payload struct {
		OutV     json.RawMessage   `json:"outV"`
		InV      json.RawMessage   `json:"inV"`
		InVs     []json.RawMessage `json:"inVs"`
		Document json.RawMessage   `json:"document"`
		Shard    json.RawMessage   `json:"shard"` // replaced `document` in 0.5.x
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return Edge{}, err
	}

	outV, err := internRaw(interner, payload.OutV)
	if err != nil {
		return nil, err
	}
	inV, err := internRaw(interner, payload.InV)
	if err != nil {
		return nil, err
	}
	document, err := internRaw(interner, payload.Document)
	if err != nil {
		return nil, err
	}

	if document == 0 {
		document, err = internRaw(interner, payload.Shard)
		if err != nil {
			return nil, err
		}
	}

	var inVs []int
	for _, inV := range payload.InVs {
		id, err := internRaw(interner, inV)
		if err != nil {
			return nil, err
		}

		inVs = append(inVs, id)
	}

	return Edge{
		OutV:     outV,
		InV:      inV,
		InVs:     inVs,
		Document: document,
	}, nil
}

// unmarshalEdgeFast attempts to unmarshal the edge without requiring use of the
// interner. Doing a bare json.Unmarshal happens is faster than unmarshalling into
// raw message and then performing strconv.Atoi.
//
// Note that we do happen to do this for edge unmarshalling. The win here comes from
// saving the of large inVs sets. Doing the same thing for element envelope identifiers
// do not net the same benefit.
func unmarshalEdgeFast(line []byte) (Edge, bool) {
	var payload struct {
		InVs     []int `json:"inVs"`
		OutV     int   `json:"outV"`
		InV      int   `json:"inV"`
		Document int   `json:"document"`
		Shard    int   `json:"shard"` // replaced `document` in 0.5.x
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return Edge{}, false
	}

	edge := Edge{
		OutV:     payload.OutV,
		InV:      payload.InV,
		InVs:     payload.InVs,
		Document: payload.Document,
	}

	if payload.Document == 0 {
		edge.Document = payload.Shard
	}

	return edge, true
}

var edgeUnmarshalers = map[string]func(line []byte) (any, error){}

var vertexUnmarshalers = map[string]func(line []byte) (any, error){
	"metaData":             unmarshalMetaData,
	"document":             unmarshalDocument,
	"documentSymbolResult": unmarshalDocumentSymbolResult,
	"range":                unmarshalRange,
	"hoverResult":          unmarshalHover,
	"moniker":              unmarshalMoniker,
	"packageInformation":   unmarshalPackageInformation,
	"diagnosticResult":     unmarshalDiagnosticResult,
}

func unmarshalMetaData(line []byte) (any, error) {
	var payload struct {
		Version     string `json:"version"`
		ProjectRoot string `json:"projectRoot"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	return MetaData{
		Version:     payload.Version,
		ProjectRoot: payload.ProjectRoot,
	}, nil
}

func unmarshalDocumentSymbolResult(line []byte) (any, error) {
	var payload struct {
		Result []*protocol.RangeBasedDocumentSymbol `json:"result"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}
	return payload.Result, nil
}

func unmarshalDocument(line []byte) (any, error) {
	var payload struct {
		URI string `json:"uri"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	return payload.URI, nil
}

func unmarshalRange(line []byte) (any, error) {
	type _position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	}
	type _range struct {
		Start _position `json:"start"`
		End   _position `json:"end"`
	}
	type _tag struct {
		FullRange *_range              `json:"fullRange,omitempty"`
		Type      string               `json:"type"`
		Text      string               `json:"text"`
		Detail    string               `json:"detail,omitempty"`
		Tags      []protocol.SymbolTag `json:"tags,omitempty"`
		Kind      int                  `json:"kind"`
	}
	var payload struct {
		Tag   *_tag     `json:"tag"`
		Start _position `json:"start"`
		End   _position `json:"end"`
	}

	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	var tag *protocol.RangeTag
	if payload.Tag != nil {
		var fullRange *protocol.RangeData
		if payload.Tag.FullRange != nil {
			fullRange = &protocol.RangeData{
				Start: protocol.Pos{
					Line:      payload.Tag.FullRange.Start.Line,
					Character: payload.Tag.FullRange.Start.Character,
				},
				End: protocol.Pos{
					Line:      payload.Tag.FullRange.End.Line,
					Character: payload.Tag.FullRange.End.Character,
				},
			}
		}
		tag = &protocol.RangeTag{
			Type:      payload.Tag.Type,
			Text:      payload.Tag.Text,
			Kind:      protocol.SymbolKind(payload.Tag.Kind),
			FullRange: fullRange,
			Detail:    payload.Tag.Detail,
			Tags:      payload.Tag.Tags,
		}
	}

	return Range{
		RangeData: protocol.RangeData{
			Start: protocol.Pos{
				Line:      payload.Start.Line,
				Character: payload.Start.Character,
			},
			End: protocol.Pos{
				Line:      payload.End.Line,
				Character: payload.End.Character,
			},
		},
		Tag: tag,
	}, nil
}

var HoverPartSeparator = "\n\n---\n\n"

func unmarshalHover(line []byte) (any, error) {
	type _hoverResult struct {
		Contents json.RawMessage `json:"contents"`
	}
	var payload struct {
		Result _hoverResult `json:"result"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	var target []json.RawMessage
	if err := unmarshaller.Unmarshal(payload.Result.Contents, &target); err != nil {
		// attempt unmarshal into either single MarkedString or MarkupContent
		v, err := unmarshalHoverPart(payload.Result.Contents)
		if err != nil {
			return nil, err
		}

		return *v, nil
	}

	var parts []string
	for _, t := range target {
		part, err := unmarshalHoverPart(t)
		if err != nil {
			return nil, err
		}

		parts = append(parts, *part)
	}

	return strings.Join(parts, HoverPartSeparator), nil
}

func unmarshalHoverPart(raw json.RawMessage) (*string, error) {
	// first, assume MarkedString or MarkupContent. This should be more likely
	var m struct {
		Kind     string
		Language string
		Value    string
	}

	err := unmarshaller.Unmarshal(raw, &m)
	if err != nil {
		// to handle the first part of the union
		// type MarkedString = string | { language: string; value: string }
		var strPayload string
		if err := unmarshaller.Unmarshal(raw, &strPayload); err == nil {
			trimmed := strings.TrimSpace(strPayload)
			return &trimmed, nil
		}
		return &strPayload, err
	}

	// now check if MarkupContent
	if m.Kind != "" {
		// TODO: validate possible values
		markup := strings.TrimSpace(protocol.NewMarkupContent(m.Value, protocol.MarkupKind(m.Kind)).String())
		return &markup, nil
	}

	// else assume MarkedString
	marked := strings.TrimSpace(protocol.NewMarkedString(m.Value, m.Language).String())

	return &marked, nil
}

func unmarshalMoniker(line []byte) (any, error) {
	var payload struct {
		Kind       string `json:"kind"`
		Scheme     string `json:"scheme"`
		Identifier string `json:"identifier"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	if payload.Kind == "" {
		payload.Kind = "local"
	}

	return Moniker{
		Kind:       payload.Kind,
		Scheme:     payload.Scheme,
		Identifier: payload.Identifier,
	}, nil
}

func unmarshalPackageInformation(line []byte) (any, error) {
	var payload struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	return PackageInformation{
		Manager: "",
		Name:    payload.Name,
		Version: payload.Version,
	}, nil
}

func unmarshalDiagnosticResult(line []byte) (any, error) {
	type _position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	}
	type _range struct {
		Start _position `json:"start"`
		End   _position `json:"end"`
	}
	type _result struct {
		Code     StringOrInt `json:"code"`
		Message  string      `json:"message"`
		Source   string      `json:"source"`
		Range    _range      `json:"range"`
		Severity int         `json:"severity"`
	}
	var payload struct {
		Results []_result `json:"result"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	var diagnostics []Diagnostic
	for _, result := range payload.Results {
		diagnostics = append(diagnostics, Diagnostic{
			Severity:       result.Severity,
			Code:           string(result.Code),
			Message:        result.Message,
			Source:         result.Source,
			StartLine:      result.Range.Start.Line,
			StartCharacter: result.Range.Start.Character,
			EndLine:        result.Range.End.Line,
			EndCharacter:   result.Range.End.Character,
		})
	}

	return diagnostics, nil
}

type StringOrInt string

func (id *StringOrInt) UnmarshalJSON(raw []byte) error {
	if raw[0] == '"' {
		var v string
		if err := unmarshaller.Unmarshal(raw, &v); err != nil {
			return err
		}

		*id = StringOrInt(v)
		return nil
	}

	var v int64
	if err := unmarshaller.Unmarshal(raw, &v); err != nil {
		return err
	}

	*id = StringOrInt(strconv.FormatInt(v, 10))
	return nil
}

// internRaw trims whitespace from the raw message and submits it to the
// interner to produce a unique identifier for this value. It is necessary
// to trim the whitespace as json-iterator can add a whitespace prefixe to
// raw messages during unmarshalling.
func internRaw(interner *Interner, raw json.RawMessage) (int, error) {
	return interner.Intern(bytes.TrimSpace(raw))
}
