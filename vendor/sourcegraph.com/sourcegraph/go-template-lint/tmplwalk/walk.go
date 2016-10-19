package tmplwalk

import "text/template/parse"

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node parse.Node) (w Visitor)
}

// Helper functions for common node types.

func walkBranchNode(v Visitor, n *parse.BranchNode) {
	Walk(v, n.Pipe)
	Walk(v, n.List)
	if n.ElseList != nil {
		Walk(v, n.ElseList)
	}
}

// Walk traverses a template parse tree in depth-first order: It
// starts by calling v.Visit(node); node must not be nil. If the
// visitor w returned by v.Visit(node) is not nil, Walk is invoked
// recursively with visitor w for each of the non-nil children of
// node, followed by a call of w.Visit(nil).
func Walk(v Visitor, node parse.Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// walk children
	switch n := node.(type) {

	case *parse.ActionNode:
		Walk(v, n.Pipe)

	case *parse.BoolNode:
		// nothing to do

	case *parse.ChainNode:
		Walk(v, n.Node)

	case *parse.CommandNode:
		for _, a := range n.Args {
			Walk(v, a)
		}

	case *parse.DotNode:
		// nothing to do

	case *parse.FieldNode:
		// nothing to do

	case *parse.IdentifierNode:
		// nothing to do

	case *parse.IfNode:
		walkBranchNode(v, &n.BranchNode)

	case *parse.ListNode:
		for _, n := range n.Nodes {
			Walk(v, n)
		}

	case *parse.NilNode:
		// nothing to do

	case *parse.NumberNode:
		// nothing to do

	case *parse.PipeNode:
		for _, d := range n.Decl {
			Walk(v, d)
		}
		for _, c := range n.Cmds {
			Walk(v, c)
		}

	case *parse.RangeNode:
		walkBranchNode(v, &n.BranchNode)

	case *parse.StringNode:
		// nothing to do

	case *parse.TemplateNode:
		if n.Pipe != nil {
			Walk(v, n.Pipe)
		}

	case *parse.TextNode:
		// nothing to do

	case *parse.VariableNode:
		// nothing to do

	case *parse.WithNode:
		walkBranchNode(v, &n.BranchNode)

	}

	v.Visit(nil)
}

type inspector func(parse.Node) bool

func (f inspector) Visit(node parse.Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

// Inspect traverses a template parse tree in depth-first order: It
// starts by calling f(node); node must not be nil. If f returns true,
// Inspect invokes f for all the non-nil children of node,
// recursively.
func Inspect(node parse.Node, f func(parse.Node) bool) {
	Walk(inspector(f), node)
}
