package api

import (
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type Builder struct {
	Document     *lsif_typed.Document
	Input        *Input
	Options      *IndexingOptions
	localCounter int
}

func (v *Builder) NewLocalSymbol(suffix string) *Symbol {
	result := NewSymbol(fmt.Sprintf("local%d-%s", v.localCounter, suffix))
	v.localCounter++
	return result
}

func (v *Builder) EmitOccurrence(symbol *Symbol, n *sitter.Node, role lsif_typed.MonikerOccurrence_Role) {
	v.Document.Occurrences = append(v.Document.Occurrences, &lsif_typed.MonikerOccurrence{
		MonikerId: symbol.Value,
		Range: &lsif_typed.Range{
			Start: &lsif_typed.Position{
				Line:      int32(n.StartPoint().Row),
				Character: int32(n.StartPoint().Column),
			},
			End: &lsif_typed.Position{
				Line:      int32(n.EndPoint().Row),
				Character: int32(n.EndPoint().Column),
			},
		},
		Role: role,
	})
}

func (v *Builder) EmitLocalOccurrence(n *sitter.Node, scope *Scope, role lsif_typed.MonikerOccurrence_Role) {
	name := NewSimpleName(v.Input.Substring(n))
	symbol := v.NewLocalSymbol(name.Value)
	scope.Bind(name, symbol)
	v.EmitOccurrence(symbol, n, role)
}

func (v *Builder) PrintlnDebug(n *sitter.Node) {
	fmt.Println(n.Type())
	fmt.Println(v.Input.Format(n))
}
