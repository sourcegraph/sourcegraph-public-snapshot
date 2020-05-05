package lsif

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

type ResultSetData struct {
	DefinitionResultID string
	ReferenceResultID  string
	HoverResultID      string
	MonikerIDs         datastructures.IDSet
}

func UnmarshalResultSetData(element Element) (ResultSetData, error) {
	return ResultSetData{MonikerIDs: datastructures.IDSet{}}, nil
}

func (d ResultSetData) SetDefinitionResultID(id string) ResultSetData {
	return ResultSetData{
		DefinitionResultID: id,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSetData) SetReferenceResultID(id string) ResultSetData {
	return ResultSetData{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  id,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSetData) SetHoverResultID(id string) ResultSetData {
	return ResultSetData{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      id,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSetData) SetMonikerIDs(ids datastructures.IDSet) ResultSetData {
	return ResultSetData{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         ids,
	}
}
