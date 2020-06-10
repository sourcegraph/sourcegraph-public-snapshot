package lsif

import "github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"

type Element struct {
	ID      string
	Type    string
	Label   string
	Payload interface{}
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

type Document struct {
	URI         string
	Contains    datastructures.IDSet
	Diagnostics datastructures.IDSet
}

type Range struct {
	StartLine          int
	StartCharacter     int
	EndLine            int
	EndCharacter       int
	DefinitionResultID string
	ReferenceResultID  string
	HoverResultID      string
	MonikerIDs         datastructures.IDSet
}

func (d Range) SetDefinitionResultID(id string) Range {
	return Range{
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

func (d Range) SetReferenceResultID(id string) Range {
	return Range{
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

func (d Range) SetHoverResultID(id string) Range {
	return Range{
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

func (d Range) SetMonikerIDs(ids datastructures.IDSet) Range {
	return Range{
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

type ResultSet struct {
	DefinitionResultID string
	ReferenceResultID  string
	HoverResultID      string
	MonikerIDs         datastructures.IDSet
}

func (d ResultSet) SetDefinitionResultID(id string) ResultSet {
	return ResultSet{
		DefinitionResultID: id,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSet) SetReferenceResultID(id string) ResultSet {
	return ResultSet{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  id,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSet) SetHoverResultID(id string) ResultSet {
	return ResultSet{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      id,
		MonikerIDs:         d.MonikerIDs,
	}
}

func (d ResultSet) SetMonikerIDs(ids datastructures.IDSet) ResultSet {
	return ResultSet{
		DefinitionResultID: d.DefinitionResultID,
		ReferenceResultID:  d.ReferenceResultID,
		HoverResultID:      d.HoverResultID,
		MonikerIDs:         ids,
	}
}

type Moniker struct {
	Kind                 string
	Scheme               string
	Identifier           string
	PackageInformationID string
}

func (d Moniker) SetPackageInformationID(id string) Moniker {
	return Moniker{
		Kind:                 d.Kind,
		Scheme:               d.Scheme,
		Identifier:           d.Identifier,
		PackageInformationID: id,
	}
}

type PackageInformation struct {
	Name    string
	Version string
}

type DiagnosticResult struct {
	Result []Diagnostic
}

type Diagnostic struct {
	Severity       int
	Code           string
	Message        string
	Source         string
	StartLine      int
	StartCharacter int
	EndLine        int
	EndCharacter   int
}
