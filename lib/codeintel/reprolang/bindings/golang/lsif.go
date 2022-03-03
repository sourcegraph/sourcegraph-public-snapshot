package golang

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
)

func (i *identifier) occurrence(roles lsiftyped.SymbolRole) *lsiftyped.Occurrence {
	return &lsiftyped.Occurrence{
		Range:       i.position.LsifRange(),
		Symbol:      i.symbol,
		SymbolRoles: int32(roles),
	}
}

func (s *reproSourceFile) symbols() []*lsiftyped.SymbolInformation {
	var result []*lsiftyped.SymbolInformation
	for _, def := range s.definitions {
		documentation := []string{"signature of " + def.name.value}
		if def.docstring != "" {
			documentation = append(documentation, def.docstring)
		}
		result = append(result, &lsiftyped.SymbolInformation{
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

func (s *reproSourceFile) occurrences() []*lsiftyped.Occurrence {
	var result []*lsiftyped.Occurrence
	for _, def := range s.definitions {
		result = append(result, def.name.occurrence(lsiftyped.SymbolRole_Definition))
		for _, ident := range def.relationIdentifiers() {
			if ident == nil {
				continue
			}
			result = append(result, ident.occurrence(lsiftyped.SymbolRole_UnspecifiedSymbolRole))
		}
	}
	for _, ref := range s.references {
		result = append(result, ref.name.occurrence(lsiftyped.SymbolRole_UnspecifiedSymbolRole))
	}
	return result
}

func (s *definitionStatement) relationships() []*lsiftyped.Relationship {
	bySymbol := map[string]*lsiftyped.Relationship{}
	for _, ident := range s.relationIdentifiers() {
		if ident == nil {
			continue
		}
		bySymbol[ident.symbol] = &lsiftyped.Relationship{Symbol: ident.symbol}
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
	var result []*lsiftyped.Relationship
	for _, value := range bySymbol {
		result = append(result, value)
	}
	return result
}
