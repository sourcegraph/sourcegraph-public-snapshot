package scip

func (r SymbolRole) Matches(occ *Occurrence) bool {
	return occ.SymbolRoles&int32(r) > 0
}