package golang

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

func (i *identifier) occurrence(roles lsif_typed.SymbolRole) *lsif_typed.Occurrence {
	return &lsif_typed.Occurrence{
		Range:       i.position.LsifRange(),
		Symbol:      i.symbol,
		SymbolRoles: int32(roles),
	}
}

func (s *reproSourceFile) symbols() []*lsif_typed.SymbolInformation {
	var result []*lsif_typed.SymbolInformation
	for _, def := range s.definitions {
		documentation := []string{"signature of " + def.name.value}
		if def.docstring != "" {
			documentation = append(documentation, def.docstring)
		}
		result = append(result, &lsif_typed.SymbolInformation{
			Symbol:        def.name.symbol,
			Documentation: documentation,
			Relationships: def.relationships(),
		})
	}
	// Ensure a stable order of relationships
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Symbol < result[j].Symbol
	})
	return result
}

func (s *reproSourceFile) occurrences() []*lsif_typed.Occurrence {
	var result []*lsif_typed.Occurrence
	for _, def := range s.definitions {
		result = append(result, def.name.occurrence(lsif_typed.SymbolRole_Definition))
		for _, ident := range def.relationIdentifiers() {
			if ident == nil {
				continue
			}
			result = append(result, ident.occurrence(lsif_typed.SymbolRole_UnspecifiedSymbolRole))
		}
	}
	for _, ref := range s.references {
		result = append(result, ref.name.occurrence(lsif_typed.SymbolRole_UnspecifiedSymbolRole))
	}
	return result
}

func (s *definitionStatement) relationships() []*lsif_typed.Relationship {
	bySymbol := map[string]*lsif_typed.Relationship{}
	for _, ident := range s.relationIdentifiers() {
		if ident == nil {
			continue
		}
		bySymbol[ident.symbol] = &lsif_typed.Relationship{Symbol: ident.symbol}
	}
	if s.implementsRelation != nil {
		bySymbol[s.implementsRelation.symbol].IsImplementation = true
	}
	if s.referencesRelation != nil {
		bySymbol[s.referencesRelation.symbol].IsReference = true
	}
	if s.typeDefinesRelation != nil {
		bySymbol[s.typeDefinesRelation.symbol].IsTypeDefinition = true
	}
	var result []*lsif_typed.Relationship
	for _, value := range bySymbol {
		result = append(result, value)
	}
	return result
}
