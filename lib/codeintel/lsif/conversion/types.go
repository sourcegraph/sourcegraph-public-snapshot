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
}

// Note [Assignment to fields of structs in maps]
//
// Go disallows `m[key].field2 = ...`. Known workarounds include:
//
// - Assign a whole struct value to the index: `m[key] = V{field1: m[key].field1, field2: ...}`
// - Change `m`'s values to be pointers
//
// This file provides convenience functions for assigning a whole struct value to the index.
//
// See https://stackoverflow.com/a/32751792/16865079

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (r Range) SetDefinitionResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     id,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (r Range) SetReferenceResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (r Range) SetImplementationResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: id,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (r Range) SetHoverResultID(id int) Range {
	return Range{
		Range:                  r.Range,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementationResultID: r.ImplementationResultID,
		HoverResultID:          id,
	}
}

type ResultSet struct {
	reader.ResultSet
	DefinitionResultID     int
	ReferenceResultID      int
	ImplementationResultID int
	HoverResultID          int
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (rs ResultSet) SetDefinitionResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     id,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (rs ResultSet) SetReferenceResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (rs ResultSet) SetImplementationResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: id,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (rs ResultSet) SetHoverResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementationResultID: rs.ImplementationResultID,
		HoverResultID:          id,
	}
}

type Moniker struct {
	reader.Moniker
	PackageInformationID int
}

// Convenience function for setting the field within a map.
//
// See Note [Assignment to fields of structs in maps]
func (m Moniker) SetPackageInformationID(id int) Moniker {
	return Moniker{
		Moniker:              m.Moniker,
		PackageInformationID: id,
	}
}
