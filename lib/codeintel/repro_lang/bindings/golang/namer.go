package repro_lang

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

func (d *reproDependency) resolveGlobalDefinitions(context *reproContext) {
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

func (d *reproSourceFile) resolveDefinitions(context *reproContext) {
	for _, def := range d.definitions {
		scope := context.globalScope
		if def.name.isLocalSymbol() {
			scope = d.localScope
		}
		symbol, ok := scope.names[def.name.value]
		if ok {
			symbol = "local ERROR_DUPLICATE_DEFINITION"
			scope = d.localScope
		} else if def.name.isLocalSymbol() {
			symbol = fmt.Sprintf("local %s", def.name.value[len("local"):])
		} else {
			symbol = newGlobalSymbol(context.pkg, d, def)
		}
		def.name.symbol = symbol
		scope.names[def.name.value] = symbol
	}
}

func (d *reproSourceFile) resolveReferences(context *reproContext) {
	for _, def := range d.definitions {
		for _, ident := range def.relationIdentifiers() {
			if ident == nil {
				continue
			}
			ident.resolveSymbol(d.localScope, context)
		}
	}
	for _, ref := range d.references {
		ref.name.resolveSymbol(d.localScope, context)
	}
}

func newGlobalSymbol(pkg *lsif_typed.Package, document *reproSourceFile, definition *definitionStatement) string {
	return fmt.Sprintf(
		"repro_lang repro_manager %v %v %v/%v",
		pkg.Name,
		pkg.Version,
		document.Source.RelativePath,
		definition.name.value,
	)
}

func newGlobalName(pkg *lsif_typed.Package, symbol *lsif_typed.Symbol) string {
	formatter := lsif_typed.DescriptorOnlyFormatter
	formatter.IncludePackageName = func(name string) bool { return name != pkg.Name }
	return "global " + formatter.FormatSymbol(symbol)
}
