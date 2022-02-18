package golang

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

// enterGlobalDefinitions inserts the names of the global symbols that are defined in this
// dependency into the provided global scope.
func (d *reproDependency) enterGlobalDefinitions(context *reproContext) {
	for _, file := range d.Sources {
		for _, definition := range file.definitions {
			if definition.name.isLocalSymbol() {
				continue
			}
			symbol := newGlobalSymbol(d.Package, file, definition)
			parsedSymbol, err := lsif_typed.ParseSymbol(symbol)
			if err != nil {
				continue
			}
			name := newGlobalName(context.pkg, parsedSymbol)
			context.globalScope.names[name] = symbol
		}
	}
}

// enterDefinitions inserts the names of the definitions into the appropriate scope (local symbols go into the local scope).
func (s *reproSourceFile) enterDefinitions(context *reproContext) {
	for _, def := range s.definitions {
		scope := context.globalScope
		if def.name.isLocalSymbol() {
			scope = s.localScope
		}
		var symbol string
		_, ok := scope.names[def.name.value]
		if ok {
			symbol = "local ERROR_DUPLICATE_DEFINITION"
		} else if def.name.isLocalSymbol() {
			symbol = fmt.Sprintf("local %s", def.name.value[len("local"):])
		} else {
			symbol = newGlobalSymbol(context.pkg, s, def)
		}
		def.name.symbol = symbol
		scope.names[def.name.value] = symbol
	}
}

// resolveReferences updates the .symbol field for all names of reference identifiers.
func (s *reproSourceFile) resolveReferences(context *reproContext) {
	for _, def := range s.definitions {
		for _, ident := range def.relationIdentifiers() {
			if ident == nil {
				continue
			}
			ident.resolveSymbol(s.localScope, context)
		}
	}
	for _, ref := range s.references {
		ref.name.resolveSymbol(s.localScope, context)
	}
}

// newGlobalSymbol returns an LSIF Typed symbol for the given definition.
func newGlobalSymbol(pkg *lsif_typed.Package, document *reproSourceFile, definition *definitionStatement) string {
	return fmt.Sprintf(
		"repro_lang repro_manager %v %v %v/%v",
		pkg.Name,
		pkg.Version,
		document.Source.RelativePath,
		definition.name.value,
	)
}

// newGlobalName returns the name of a symbol that is used to query the scope.
func newGlobalName(pkg *lsif_typed.Package, symbol *lsif_typed.Symbol) string {
	formatter := lsif_typed.DescriptorOnlyFormatter
	formatter.IncludePackageName = func(name string) bool { return name != pkg.Name }
	return "global " + formatter.FormatSymbol(symbol)
}
