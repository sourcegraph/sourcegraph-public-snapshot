package conversion

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"

type Element reader.Element
type Edge reader.Edge
type MetaData reader.MetaData
type PackageInformation reader.PackageInformation
type Diagnostic reader.Diagnostic

type Range struct {
	reader.Range
	DefinitionResultID     int
	ReferenceResultID      int
	ImplementationResultID int
	HoverResultID          int
	DocumentationResultID  int
}

func (r Range) SetDefinitionResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     id,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          r.HoverResultID,
		DocumentationResultID:  r.DocumentationResultID,
	}
}

func (r Range) SetReferenceResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          r.HoverResultID,
		DocumentationResultID:  r.DocumentationResultID,
	}
}

func (r Range) SetImplementationResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: id,
		HoverResultID:          r.HoverResultID,
		DocumentationResultID:  r.DocumentationResultID,
	}
}

func (r Range) SetHoverResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          id,
		DocumentationResultID:  r.DocumentationResultID,
	}
}

func (r Range) SetDocumentationResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          r.HoverResultID,
		DocumentationResultID:  id,
	}
}

type ResultSet struct {
	reader.ResultSet
	DefinitionResultID     int
	ReferenceResultID      int
	ImplementationResultID int
	HoverResultID          int
	DocumentationResultID  int
}

func (rs ResultSet) SetDefinitionResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     id,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          rs.HoverResultID,
		DocumentationResultID:  rs.DocumentationResultID,
	}
}

func (rs ResultSet) SetReferenceResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          rs.HoverResultID,
		DocumentationResultID:  rs.DocumentationResultID,
	}
}

func (rs ResultSet) SetImplementationResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: id,
		HoverResultID:          rs.HoverResultID,
		DocumentationResultID:  rs.DocumentationResultID,
	}
}

func (rs ResultSet) SetHoverResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          id,
		DocumentationResultID:  rs.DocumentationResultID,
	}
}

func (rs ResultSet) SetDocumentationResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          rs.HoverResultID,
		DocumentationResultID:  id,
	}
}

type Moniker struct {
	reader.Moniker
	PackageInformationID int
}

func (m Moniker) SetPackageInformationID(id int) Moniker {
	return Moniker{
		Moniker:              m.Moniker,
		PackageInformationID: id,
	}
}
