package repro_lang

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type definitionStatement struct {
	docstring           string
	name                *identifier
	implementsRelation  *identifier
	referencesRelation  *identifier
	typeDefinesRelation *identifier
}

func (s *definitionStatement) relationIdentifiers() []*identifier {
	return []*identifier{s.implementsRelation, s.referencesRelation, s.typeDefinesRelation}
}

type referenceStatement struct {
	name *identifier
}

type identifier struct {
	value    string
	symbol   string
	position *lsif_typed.RangePosition
}

func (i *identifier) resolveSymbol(localScope *reproScope, context *reproContext) {
	scope := context.globalScope
	if i.isLocalSymbol() {
		scope = localScope
	}
	symbol, ok := scope.names[i.value]
	if !ok {
		if i.value == "global global-workspace hello.repro/hello()." {
			fmt.Println("scope", context.globalScope)
		}
		symbol = "local ERROR_UNRESOLVED_SYMBOL"
		scope = localScope
	}
	i.symbol = symbol
}

func (i *identifier) isLocalSymbol() bool {
	return strings.HasPrefix(i.value, "local")
}
