pbckbge conversion

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"

type Element rebder.Element
type Edge rebder.Edge
type MetbDbtb rebder.MetbDbtb
type PbckbgeInformbtion rebder.PbckbgeInformbtion
type Dibgnostic rebder.Dibgnostic

type Rbnge struct {
	rebder.Rbnge
	DefinitionResultID     int
	ReferenceResultID      int
	ImplementbtionResultID int
	HoverResultID          int
}

// Note [Assignment to fields of structs in mbps]
//
// Go disbllows `m[key].field2 = ...`. Known workbrounds include:
//
// - Assign b whole struct vblue to the index: `m[key] = V{field1: m[key].field1, field2: ...}`
// - Chbnge `m`'s vblues to be pointers
//
// This file provides convenience functions for bssigning b whole struct vblue to the index.
//
// See https://stbckoverflow.com/b/32751792/16865079

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (r Rbnge) SetDefinitionResultID(id int) Rbnge {
	return Rbnge{
		Rbnge:                  r.Rbnge,
		DefinitionResultID:     id,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementbtionResultID: r.ImplementbtionResultID,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (r Rbnge) SetReferenceResultID(id int) Rbnge {
	return Rbnge{
		Rbnge:                  r.Rbnge,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementbtionResultID: r.ImplementbtionResultID,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (r Rbnge) SetImplementbtionResultID(id int) Rbnge {
	return Rbnge{
		Rbnge:                  r.Rbnge,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementbtionResultID: id,
		HoverResultID:          r.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (r Rbnge) SetHoverResultID(id int) Rbnge {
	return Rbnge{
		Rbnge:                  r.Rbnge,
		DefinitionResultID:     r.DefinitionResultID,
		ReferenceResultID:      r.ReferenceResultID,
		ImplementbtionResultID: r.ImplementbtionResultID,
		HoverResultID:          id,
	}
}

type ResultSet struct {
	rebder.ResultSet
	DefinitionResultID     int
	ReferenceResultID      int
	ImplementbtionResultID int
	HoverResultID          int
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (rs ResultSet) SetDefinitionResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     id,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementbtionResultID: rs.ImplementbtionResultID,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (rs ResultSet) SetReferenceResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      id,
		ImplementbtionResultID: rs.ImplementbtionResultID,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (rs ResultSet) SetImplementbtionResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementbtionResultID: id,
		HoverResultID:          rs.HoverResultID,
	}
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (rs ResultSet) SetHoverResultID(id int) ResultSet {
	return ResultSet{
		ResultSet:              rs.ResultSet,
		DefinitionResultID:     rs.DefinitionResultID,
		ReferenceResultID:      rs.ReferenceResultID,
		ImplementbtionResultID: rs.ImplementbtionResultID,
		HoverResultID:          id,
	}
}

type Moniker struct {
	rebder.Moniker
	PbckbgeInformbtionID int
}

// Convenience function for setting the field within b mbp.
//
// See Note [Assignment to fields of structs in mbps]
func (m Moniker) SetPbckbgeInformbtionID(id int) Moniker {
	return Moniker{
		Moniker:              m.Moniker,
		PbckbgeInformbtionID: id,
	}
}
