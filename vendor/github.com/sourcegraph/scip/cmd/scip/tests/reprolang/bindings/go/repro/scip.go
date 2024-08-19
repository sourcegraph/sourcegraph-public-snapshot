package repro

import (
	"sort"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

func (i *identifier) occurrence(roles scip.SymbolRole) *scip.Occurrence {
	var diagnostics []*scip.Diagnostic
	if strings.HasPrefix(i.value, "deprecated") {
		diagnostics = []*scip.Diagnostic{{
			Severity: scip.Severity_Warning,
			Message: "deprecated identifier",
		}}
	}

	return &scip.Occurrence{
		Range:       i.position.SCIPRange(),
		Symbol:      i.symbol,
		SymbolRoles: int32(roles),
		Diagnostics: diagnostics,
	}
}

func (s *reproSourceFile) symbols() []*scip.SymbolInformation {
	var result []*scip.SymbolInformation
	for _, rel := range s.relationships {
		result = append(result, &scip.SymbolInformation{
			Symbol:        rel.name.symbol,
			Documentation: nil,
			Relationships: rel.relations.toSCIP(),
		})
	}
	for _, def := range s.definitions {
		if strings.Index(def.name.value, "NoSymbolInformation") >= 0 {
			continue
		}
		documentation := []string{"signature of " + def.name.value}
		if def.docstring != "" {
			documentation = append(documentation, def.docstring)
		}
		result = append(result, &scip.SymbolInformation{
			Symbol:        def.name.symbol,
			Documentation: documentation,
			Relationships: def.relations.toSCIP(),
		})
	}
	// Ensure a stable order of relationships
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Symbol < result[j].Symbol
	})
	return result
}

func (s *reproSourceFile) occurrences() []*scip.Occurrence {
	var result []*scip.Occurrence
	emit := func(rel relationships) {
		for _, ident := range rel.identifiers() {
			if ident == nil {
				continue
			}
			result = append(result, ident.occurrence(scip.SymbolRole_UnspecifiedSymbolRole))
		}
	}
	for _, def := range s.definitions {
		result = append(result, def.name.occurrence(scip.SymbolRole_Definition))
		emit(def.relations)
	}
	for _, rel := range s.relationships {
		emit(rel.relations)
	}
	for _, ref := range s.references {
		role := scip.SymbolRole_UnspecifiedSymbolRole
		if ref.isForwardDef {
			role = scip.SymbolRole_ForwardDefinition
		}
		result = append(result, ref.name.occurrence(role))
	}
	return result
}

func (r *relationships) toSCIP() []*scip.Relationship {
	bySymbol := map[string]*scip.Relationship{}
	for _, ident := range r.identifiers() {
		if ident == nil {
			continue
		}
		bySymbol[ident.symbol] = &scip.Relationship{Symbol: ident.symbol}
	}
	if r.implementsRelation != nil {
		bySymbol[r.implementsRelation.symbol].IsImplementation = true
	}
	if r.referencesRelation != nil {
		bySymbol[r.referencesRelation.symbol].IsReference = true
	}
	if r.typeDefinesRelation != nil {
		bySymbol[r.typeDefinesRelation.symbol].IsTypeDefinition = true
	}
	if r.definedByRelation != nil {
		bySymbol[r.definedByRelation.symbol].IsDefinition = true
	}
	var result []*scip.Relationship
	for _, value := range bySymbol {
		result = append(result, value)
	}
	return result
}
