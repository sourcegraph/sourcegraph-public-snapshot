package lsif

import (
	"encoding/json"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

type RangeData struct {
	StartLine          int                  `json:"startLine"`
	StartCharacter     int                  `json:"startCharacter"`
	EndLine            int                  `json:"endLine"`
	EndCharacter       int                  `json:"endCharacter"`
	DefinitionResultID string               `json:"definitionResultId"`
	ReferenceResultID  string               `json:"referenceResultId"`
	HoverResultID      string               `json:"hoverResultId"`
	MonikerIDs         datastructures.IDSet `json:"monikerIds"`
}

func UnmarshalRangeData(element Element) (RangeData, error) {
	type Position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	}

	type RangeVertex struct {
		Start Position `json:"start"`
		End   Position `json:"end"`
	}

	var payload RangeVertex
	if err := json.Unmarshal(element.Raw, &payload); err != nil {
		return RangeData{}, err
	}

	return RangeData{
		StartLine:      payload.Start.Line,
		StartCharacter: payload.Start.Character,
		EndLine:        payload.End.Line,
		EndCharacter:   payload.End.Character,
		MonikerIDs:     datastructures.IDSet{},
	}, nil
}

func (d RangeData) SetDefinitionResultID(id string) RangeData {
	return RangeData{
		StartLine:          d.StartLine,
		StartCharacter:     d.StartCharacter,
		EndLine:            d.EndLine,
		EndCharacter:       d.EndCharacter,
		DefinitionResultID: id,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d RangeData) SetReferenceResultID(id string) RangeData {
	return RangeData{
		StartLine:          d.StartLine,
		StartCharacter:     d.StartCharacter,
		EndLine:            d.EndLine,
		EndCharacter:       d.EndCharacter,
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  id,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d RangeData) SetHoverResultID(id string) RangeData {
	return RangeData{
		StartLine:          d.StartLine,
		StartCharacter:     d.StartCharacter,
		EndLine:            d.EndLine,
		EndCharacter:       d.EndCharacter,
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      id,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d RangeData) SetMonikerIDs(ids datastructures.IDSet) RangeData {
	return RangeData{
		StartLine:          d.StartLine,
		StartCharacter:     d.StartCharacter,
		EndLine:            d.EndLine,
		EndCharacter:       d.EndCharacter,
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         ids,
	}
}
