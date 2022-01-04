package api

import sitter "github.com/smacker/go-tree-sitter"

type Scope struct {
	Outer    *Scope
	Bindings []*binding
	Node     *sitter.Node
}

type binding struct {
	Name   *Name
	Symbol *Symbol
}

func NewScope() *Scope {
	return &Scope{}
}

func (s *Scope) NewInnerScope() *Scope {
	newScope := NewScope()
	newScope.Outer = s
	return newScope
}

func (s *Scope) Bind(name *Name, symbol *Symbol) {
	s.Bindings = append(s.Bindings, &binding{Name: name, Symbol: symbol})
}

func (s *Scope) Lookup(name *Name) *Symbol {
	scope := s
	for scope != nil {
		for _, b := range scope.Bindings {
			if b.Name.IsEqual(name) {
				return b.Symbol
			}
		}
		scope = scope.Outer
	}
	return nil
}
