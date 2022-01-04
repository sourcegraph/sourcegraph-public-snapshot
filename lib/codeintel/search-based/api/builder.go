package api

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type Builder struct {
	Document     *lsif_typed.Document
	Input        *Input
	Options      *IndexingOptions
	Cursor       *sitter.TreeCursor
	Types        []string
	FieldNames   []string
	Scope        *Scope
	localCounter int
}

type LocalIntelGrammar struct {
	Identifiers  map[string]struct{}
	Fingerprints []DefinitionFingerprint
}

type DefinitionFingerprint struct {
	ParentTypes      []string
	ParentFieldNames []string
}

func Index(
	ctx context.Context,
	input *Input,
	language *sitter.Language,
	grammar LocalIntelGrammar,
) (*lsif_typed.Document, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(language)
	tree, err := parser.ParseCtx(ctx, nil, input.Bytes)
	if err != nil {
		return nil, err
	}
	doc := &lsif_typed.Document{
		Uri:         input.Uri(),
		Occurrences: nil,
	}
	cursor := sitter.NewTreeCursor(tree.RootNode())
	visitor := &Builder{
		Document: doc,
		Input:    input,
		Cursor:   cursor,
		Scope:    NewScope(),
	}
	for visitor.NextNode() {
		node := visitor.Cursor.CurrentNode()
		var definitionFingerprint *DefinitionFingerprint
		for _, fingerprint := range grammar.Fingerprints {
			for i, parentType := range fingerprint.ParentTypes {
				if i >= len(visitor.Types) {
					break
				}
				j := len(visitor.Types) - i - 1
				if parentType != visitor.Types[j] {
					break
				}
				if i < len(fingerprint.ParentFieldNames) &&
					fingerprint.ParentFieldNames[i] != visitor.FieldNames[j] {
					break
				}
				if i == len(fingerprint.ParentTypes)-1 {
					definitionFingerprint = &fingerprint
				}
			}
			if definitionFingerprint != nil {
				break
			}
		}
		if definitionFingerprint != nil {
			scope := visitor.Scope
			for i := 0; i < len(definitionFingerprint.ParentTypes); i++ {
				if definitionFingerprint.ParentTypes[i] != scope.Node.Type() {
					panic(fmt.Sprintf(
						"mis-matching parent type: fingerprint.Parent %v scope.Type %v",
						definitionFingerprint.ParentTypes[i],
						scope.Node.Type(),
					))
				}
				scope = scope.Outer
			}
			visitor.EmitLocalOccurrence(
				node,
				scope,
				lsif_typed.MonikerOccurrence_ROLE_DEFINITION,
			)
		} else if _, ok := grammar.Identifiers[node.Type()]; ok {
			name := NewSimpleName(input.Substring(node))
			sym := visitor.Scope.Lookup(name)
			if sym != nil {
				visitor.EmitOccurrence(sym, node, lsif_typed.MonikerOccurrence_ROLE_REFERENCE)
			}
		}
	}
	return visitor.Document, nil
}

func (b *Builder) popNode() {
	n := len(b.Types)
	b.Types = b.Types[0 : n-1]
	b.FieldNames = b.FieldNames[0 : n-1]
	b.Scope = b.Scope.Outer
}

func (b *Builder) pushNode() {
	b.Types = append(b.Types, b.Cursor.CurrentNode().Type())
	b.FieldNames = append(b.FieldNames, b.Cursor.CurrentFieldName())
	b.Scope = b.Scope.NewInnerScope()
	b.Scope.Node = b.Cursor.CurrentNode()
}

func (b *Builder) NextNode() bool {
	for b.nextAnyNode() {
		//fieldName := b.Cursor.CurrentFieldName()
		b.pushNode()
		return true
	}
	return false
}

func (b *Builder) nextAnyNode() bool {
	isFirstChild := b.Cursor.GoToFirstChild()
	if isFirstChild {
		return true
	}
	isNextSibling := b.Cursor.GoToNextSibling()
	if isNextSibling {
		b.popNode()
		return true
	}
	for b.Cursor.GoToParent() {
		b.popNode()
		isNextSibling = b.Cursor.GoToNextSibling()
		if isNextSibling {
			b.popNode()
			return true
		}
	}
	return false
}

func (b *Builder) NewLocalSymbol(suffix string) *Symbol {
	result := NewSymbol(fmt.Sprintf("local%d-%s", b.localCounter, suffix))
	b.localCounter++
	return result
}

func (b *Builder) EmitOccurrence(symbol *Symbol, n *sitter.Node, role lsif_typed.MonikerOccurrence_Role) {
	b.Document.Occurrences = append(b.Document.Occurrences, &lsif_typed.MonikerOccurrence{
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

func (b *Builder) EmitLocalOccurrence(n *sitter.Node, scope *Scope, role lsif_typed.MonikerOccurrence_Role) {
	name := NewSimpleName(b.Input.Substring(n))
	symbol := b.NewLocalSymbol(name.Value)
	scope.Bind(name, symbol)
	b.EmitOccurrence(symbol, n, role)
}

func (b *Builder) PrintlnDebug(n *sitter.Node) {
	fmt.Println(n.Type())
	fmt.Println(b.Input.Format(n))
}
