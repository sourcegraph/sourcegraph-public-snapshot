package lsif

import "github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"

type Element struct {
	ID      string
	Type    string
	Label   string
	Element interface{} // TODO(efritz) - rename
}

type Edge struct {
	OutV     string
	InV      string
	InVs     []string
	Document string
}

type MetaData struct {
	Version     string
	ProjectRoot string
}

type DocumentData struct {
	URI      string
	Contains datastructures.IDSet
}

// TODO(efritz) - rename all *Data structs
type RangeData struct {
	StartLine          int
	StartCharacter     int
	EndLine            int
	EndCharacter       int
	DefinitionResultID string
	ReferenceResultID  string
	HoverResultID      string
	MonikerIDs         datastructures.IDSet
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

type ResultSetData struct {
	DefinitionResultID string
	ReferenceResultID  string
	HoverResultID      string
	MonikerIDs         datastructures.IDSet
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

type MonikerData struct {
	Kind                 string
	Scheme               string
	Identifier           string
	PackageInformationID string
}

func (d MonikerData) SetPackageInformationID(id string) MonikerData {
	return MonikerData{
		Kind:                 d.Kind,
		Scheme:               d.Scheme,
		Identifier:           d.Identifier,
		PackageInformationID: id,
	}
}

type PackageInformationData struct {
	Name    string
	Version string
}
