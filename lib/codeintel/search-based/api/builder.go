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
	Names        []string
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
	}
	for visitor.NextNode() {
		var definitionFingerprint *DefinitionFingerprint
		for _, fingerprint := range grammar.Fingerprints {
			if definitionFingerprint != nil {
				break
			}
			for i, parentType := range fingerprint.ParentTypes {
				if i >= len(visitor.Types) {
					break
				}
				j := len(visitor.Types) - i - 1
				if parentType != visitor.Types[j] {
					break
				}
				if i < len(fingerprint.ParentFieldNames) &&
					fingerprint.ParentFieldNames[i] != visitor.Names[j] {
					break
				}
				if i == len(fingerprint.ParentTypes)-1 {
					definitionFingerprint = &fingerprint
				}
			}
		}
		node := visitor.Cursor.CurrentNode()
		if definitionFingerprint != nil {
			scope := visitor.Scope
			for i := 0; i < len(definitionFingerprint.ParentTypes) && scope.Outer != nil; i++ {
				scope = scope.Outer
			}
			visitor.EmitLocalOccurrence(
				node,
				scope,
				lsif_typed.MonikerOccurrence_ROLE_DEFINITION,
			)
		} else if _, ok := grammar.Identifiers[node.Type()]; ok {
			sym := visitor.Scope.Lookup(NewSimpleName(visitor.Input.Substring(node)))
			if sym != nil {
				visitor.EmitOccurrence(sym, node, lsif_typed.MonikerOccurrence_ROLE_REFERENCE)
			}
		}
		//if node.Type() == "identifier" {
		//	fmt.Println(visitor.Types)
		//	fmt.Println(visitor.Names)
		//	fmt.Println(input.Format(node))
		//}
	}
	return visitor.Document, nil
}

func (b *Builder) popName() {
	n := len(b.Types)
	if n > 0 {
		b.Types = b.Types[0 : n-1]
		b.Names = b.Names[0 : n-1]
		b.Scope = b.Scope.Outer
	}
}

func (b *Builder) pushName() {
	b.Types = append(b.Types, b.Cursor.CurrentNode().Type())
	b.Names = append(b.Names, b.Cursor.CurrentFieldName())
	b.Scope = b.Scope.NewInnerScope()
}

func (b *Builder) NextNode() bool {
	for b.nextAnyNode() {
		//fieldName := b.Cursor.CurrentFieldName()
		b.pushName()
		return true
		//if fieldName != "" {
		//}
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
		b.popName()
		return true
	}
	for b.Cursor.GoToParent() {
		b.popName()
		isNextSibling = b.Cursor.GoToNextSibling()
		if isNextSibling {
			b.popName()
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
